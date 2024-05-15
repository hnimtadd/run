"""
this package includes sdk that help you run the handler in the run runtime
"""

import json
import sys
from typing import Any, Callable, Dict
import struct
from .types.http_request import HTTPRequest
from .types.http_response import HTTPResponse

magic_len = 4


def lambda_handler(h: Callable[[HTTPRequest], HTTPResponse]) -> None:
    """
    LambdaHandler could wrap your handler then pass to the run runtime
    """
    sys.stdout.flush()

    # read bytes from std in the parse to HTTPRequest
    b = sys.stdin.buffer.read()

    inp: Dict[str, Any] = json.loads(b.decode())
    req: HTTPRequest = HTTPRequest(
        Body=inp.get("body", ""),
        Method=inp.get("method", ""),
        Url=inp.get("url", ""),
        EndpointId=inp.get("endpoint_id", ""),
        Env=inp.get("env", {}),
        Header=inp.get("header", {}),
        Runtime=inp.get("runtime", ""),
        DeploymentId=inp.get("deployment_id", ""),
        Id=inp.get("id", ""),
    )
    res = h(req)

    oup = {
        "body": res.Body,
        "code": res.Code,
        "request_id": req.Id,
        "header": res.Header,
    }
    body_bytes = json.dumps(oup).encode("utf-8")

    # write response information to sandbox stdout, using for check valid
    # response, currently, using json object instead of protobuf
    written = sys.stdout.buffer.write(body_bytes)
    if written != len(body_bytes):
        raise Exception(
            f"given bytes with length: {len(body_bytes)}, written: {written}"
        )
    sys.stdout.flush()

    body_length = len(body_bytes)
    # convert to uint16 since the runtime are treated this length as uint16
    body_length_bytes = struct.pack("<H", body_length)

    written = sys.stdout.buffer.write(body_length_bytes)
    if written != len(body_length_bytes):
        raise Exception(
            f"given bytes with length: {len(body_bytes)}, written: {written}"
        )
    sys.stdout.flush()
