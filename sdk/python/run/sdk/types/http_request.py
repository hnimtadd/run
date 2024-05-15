from dataclasses import dataclass
from typing import Dict, List


@dataclass
class HTTPRequest:
    Body: str
    Method: str
    Url: str
    EndpointId: str
    Env: Dict[str, str]
    Header: Dict[str, List[str]]
    Runtime: str
    DeploymentId: str
    Id: str
