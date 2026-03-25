package com.openpay.sdk.exceptions;

public class OpenPayException extends RuntimeException {
    public OpenPayException(String message) {
        super(message);
    }

    public OpenPayException(String message, Throwable cause) {
        super(message, cause);
    }
}
