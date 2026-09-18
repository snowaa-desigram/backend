<?php
// GENERATED CODE -- DO NOT EDIT!

namespace Desigram\Ping\V1;

/**
 * Заглушка Go-микросервиса: проверка связи core -> go.
 */
class PingServiceClient extends \Grpc\BaseStub {

    /**
     * @param string $hostname hostname
     * @param array $opts channel options
     * @param \Grpc\Channel $channel (optional) re-use channel object
     */
    public function __construct($hostname, $opts, $channel = null) {
        parent::__construct($hostname, $opts, $channel);
    }

    /**
     * @param \Desigram\Ping\V1\PingRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Ping(\Desigram\Ping\V1\PingRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/desigram.ping.v1.PingService/Ping',
        $argument,
        ['\Desigram\Ping\V1\PingResponse', 'decode'],
        $metadata, $options);
    }

}
