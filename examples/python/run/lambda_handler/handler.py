from run.sdk import lambda_handler
from run.sdk.types.http_response import HTTPResponse
from run.sdk.types.http_request import HTTPRequest


def handler(req: HTTPRequest) -> HTTPResponse:
    print("hello, this is a log", flush=True)
    print("receive request", req, flush=True)
    return HTTPResponse(
        Body="hello from the index pages",
        Code=200,
        RequestId="id",
        Header={"Content-Type": list(["application/json"])},
    )


lambda_handler(handler)
