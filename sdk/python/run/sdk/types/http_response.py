from typing import Dict, List
from dataclasses import dataclass


@dataclass
class HTTPResponse:
    Body: str
    Code: int
    RequestId: str
    Header: Dict[str, List[str]]
