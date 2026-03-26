"""Error types for the Open Pay SDK."""


class OpenPayError(Exception):
    """Base error class for Open Pay SDK errors."""
    pass


class AuthenticationError(OpenPayError):
    """Raised when API authentication fails."""
    pass


class APIError(OpenPayError):
    """Raised when the API returns an error response."""

    def __init__(self, code: str, message: str, status: int):
        self.code = code
        self.status = status
        super().__init__(f"{message} ({code}, HTTP {status})")
