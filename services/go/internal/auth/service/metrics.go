package service

import (
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/metric"

	"github.com/snowaa-desigram/backend/services/go/internal/auth/store"
)

// Бизнес-метрики auth. HTTP-метрики (http_server_requests_*) go-zero rest снимает сам;
// здесь — что именно произошло с точки зрения продукта: сколько регистраций, логинов, отказов и почему.
// Экспорт — тот же /metrics на Prometheus.Port (etc/auth.yaml); в тестах prometheus выключен и вызовы — no-op.
var (
	// auth_operations_total{op, result}: op — метод Service, result — ok | код ошибки из common.yaml | internal.
	opsTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "auth",
		Name:      "operations_total",
		Help:      "auth operations by outcome",
		Labels:    []string{"op", "result"},
	})
	// auth_operation_duration_ms{op}: сколько стоит операция целиком (bcrypt + БД + SMTP).
	opsDuration = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: "auth",
		Name:      "operation_duration_ms",
		Help:      "auth operation duration (ms)",
		Labels:    []string{"op"},
		Buckets:   []float64{10, 50, 100, 250, 500, 1000, 2500, 5000},
	})
	// auth_mail_total{purpose, result} и auth_mail_duration_ms{purpose}: отправка кодов по SMTP.
	mailTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "auth",
		Name:      "mail_total",
		Help:      "verification mails by outcome",
		Labels:    []string{"purpose", "result"},
	})
	mailDuration = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: "auth",
		Name:      "mail_duration_ms",
		Help:      "SMTP send duration (ms)",
		Labels:    []string{"purpose"},
		Buckets:   []float64{10, 50, 100, 250, 500, 1000, 2500, 5000},
	})
)

// track — defer в каждом методе Service: `defer track("login", &err)()`.
func track(op string, err *error) func() {
	start := time.Now()
	return func() {
		opsDuration.Observe(time.Since(start).Milliseconds(), op)
		opsTotal.Inc(op, ResultLabel(*err))
	}
}

func trackMail(purpose store.CodePurpose, start time.Time, err error) {
	mailDuration.Observe(time.Since(start).Milliseconds(), string(purpose))
	result := "ok"
	if err != nil {
		result = "error"
	}
	mailTotal.Inc(string(purpose), result)
}

// ResultLabel: nil → ok, *Error → его код (ограниченное множество из common.yaml), иное → internal.
func ResultLabel(err error) string {
	if err == nil {
		return "ok"
	}
	var e *Error
	if errors.As(err, &e) {
		return string(e.Code)
	}
	return "internal"
}
