"""
this package includes sdk that help you run the handler in the run runtime
"""

import json
import sys
from typing import Any, Callable, Dict


def lambda_handler(h: Callable[[Dict[str, Any]], Dict[str, Any]]) -> None:
    """
    LambdaHandler could wrap your handler then pass to the run runtime
    """

    # read bytes from std in the parse to HTTPRequest
    b = sys.stdin.buffer.read()
    inp: Dict[str, Any] = json.loads(b.decode())
    print("input:", inp)
    proto_request = {
        "body": inp.get("body"),
        "method": inp.get("method"),
        "url": inp.get("url"),
        "endpoint_id": inp.get("endpoint_id"),
        "env": inp.get("env"),
        "header": inp.get("header"),
        "runtime": inp.get("runtime"),
        "deployment_id": inp.get("deployment_id"),
        "id": inp.get("id"),
    }
    req: Dict[str, Any] = proto_request
    res = h(req)
    if sys.stderr.buffer.readable():
        log_bytes = sys.stderr.buffer.read().decode("utf-8")
        # write log from stderr to the stdout
        sys.stdout.buffer.write(log_bytes.encode("utf-8"))

    oup = {
        "body": res.get("body"),
        "code": res.get("code"),
        "request_id": proto_request.get("id"),
        "header": proto_request.get("header"),
    }

    body_bytes = json.dumps(oup).encode("utf-8")

    # write response information to sandbox stdout, using for check valid
    # response, currently, using json object instead of protobuf
    sys.stdout.buffer.write(body_bytes)

    sys.stdout.buffer.write(len(body_bytes).to_bytes(4, "little"))
