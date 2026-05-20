class VNodeXError(Exception):
    """Base error for the vnodex Python client."""


class VNodeXHTTPError(VNodeXError):
    """Raised when the node returns a non-2xx HTTP response."""

    def __init__(self, status_code: int, message: str, body: str | None = None):
        self.status_code = status_code
        self.message = message
        self.body = body
        detail = f"{status_code} {message}"
        if body:
            detail = f"{detail}: {body}"
        super().__init__(detail)