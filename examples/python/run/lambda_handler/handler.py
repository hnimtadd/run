from typing import Dict, Any

from run.sdk import lambda_handler


def handler(req: Dict[str, Any]) -> Dict[str, Any]:
    return {"msg": "Hello world from python handler"}


lambda_handler(handler)
