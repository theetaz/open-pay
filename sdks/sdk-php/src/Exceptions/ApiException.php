<?php

declare(strict_types=1);

namespace OpenPay\Exceptions;

class ApiException extends OpenPayException
{
    public function __construct(
        public readonly string $errorCode,
        string $message,
        public readonly int $httpStatus,
    ) {
        parent::__construct("{$message} ({$errorCode}, HTTP {$httpStatus})", $httpStatus);
    }
}
