"""Бутстрап gRPC-сервера: health, reflection, логирование, graceful shutdown.

Использование:
    from desigram_common import Settings, serve

    settings = Settings(name="telegram")
    serve(settings, register=lambda s: add_XServicer_to_server(XServicer(), s),
          service_names=[x_pb2.DESCRIPTOR.services_by_name["X"].full_name])
"""

import logging
import signal
import threading
from collections.abc import Callable, Iterable
from concurrent import futures

import grpc
from grpc_health.v1 import health, health_pb2, health_pb2_grpc
from grpc_reflection.v1alpha import reflection

from desigram_common.settings import Settings

log = logging.getLogger("desigram")


class _LoggingInterceptor(grpc.ServerInterceptor):
    def intercept_service(self, continuation, handler_call_details):
        log.info("rpc %s", handler_call_details.method)
        return continuation(handler_call_details)


def serve(
    settings: Settings,
    register: Callable[[grpc.Server], None],
    service_names: Iterable[str] = (),
) -> None:
    logging.basicConfig(
        level=settings.log_level,
        format="%(asctime)s %(levelname)s %(name)s: %(message)s",
    )

    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=settings.max_workers),
        interceptors=[_LoggingInterceptor()],
    )
    register(server)

    health_servicer = health.HealthServicer()
    health_pb2_grpc.add_HealthServicer_to_server(health_servicer, server)
    health_servicer.set("", health_pb2.HealthCheckResponse.SERVING)

    reflection.enable_server_reflection(
        (*service_names, reflection.SERVICE_NAME, health.SERVICE_NAME),
        server,
    )

    server.add_insecure_port(f"[::]:{settings.port}")
    server.start()
    log.info("%s: grpc listening on :%d", settings.name, settings.port)

    stop = threading.Event()

    def _on_signal(signum, _frame):
        log.info("%s: got %s, shutting down", settings.name, signal.Signals(signum).name)
        health_servicer.set("", health_pb2.HealthCheckResponse.NOT_SERVING)
        stop.set()

    signal.signal(signal.SIGTERM, _on_signal)
    signal.signal(signal.SIGINT, _on_signal)

    stop.wait()
    server.stop(settings.shutdown_grace).wait()
