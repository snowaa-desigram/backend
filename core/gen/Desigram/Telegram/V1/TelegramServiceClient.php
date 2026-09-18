<?php
// GENERATED CODE -- DO NOT EDIT!

namespace Desigram\Telegram\V1;

/**
 * Python-микросервис Telegram: отправка изображений и т.д.
 */
class TelegramServiceClient extends \Grpc\BaseStub {

    /**
     * @param string $hostname hostname
     * @param array $opts channel options
     * @param \Grpc\Channel $channel (optional) re-use channel object
     */
    public function __construct($hostname, $opts, $channel = null) {
        parent::__construct($hostname, $opts, $channel);
    }

    /**
     * @param \Desigram\Telegram\V1\SendPhotoRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function SendPhoto(\Desigram\Telegram\V1\SendPhotoRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/desigram.telegram.v1.TelegramService/SendPhoto',
        $argument,
        ['\Desigram\Telegram\V1\SendPhotoResponse', 'decode'],
        $metadata, $options);
    }

}
