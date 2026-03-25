package com.openpay.sdk.exceptions;

public class ApiException extends OpenPayException {
    private final String code;
    private final int httpStatus;

    public ApiException(String code, String message, int httpStatus) {
        super(message + " (" + code + ", HTTP " + httpStatus + ")");
        this.code = code;
        this.httpStatus = httpStatus;
    }

    public String getCode() { return code; }
    public int getHttpStatus() { return httpStatus; }
    public boolean isNotFound() { return httpStatus == 404; }
    public boolean isAuthError() { return httpStatus == 401; }
}
