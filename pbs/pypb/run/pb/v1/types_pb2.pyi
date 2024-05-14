from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class HTTPRequest(_message.Message):
    __slots__ = ("body", "method", "url", "endpoint_id", "env", "header", "runtime", "deployment_id", "id")
    class EnvEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    class HeaderEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: HeaderFields
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[HeaderFields, _Mapping]] = ...) -> None: ...
    BODY_FIELD_NUMBER: _ClassVar[int]
    METHOD_FIELD_NUMBER: _ClassVar[int]
    URL_FIELD_NUMBER: _ClassVar[int]
    ENDPOINT_ID_FIELD_NUMBER: _ClassVar[int]
    ENV_FIELD_NUMBER: _ClassVar[int]
    HEADER_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_ID_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    body: bytes
    method: str
    url: str
    endpoint_id: str
    env: _containers.ScalarMap[str, str]
    header: _containers.MessageMap[str, HeaderFields]
    runtime: str
    deployment_id: str
    id: str
    def __init__(self, body: _Optional[bytes] = ..., method: _Optional[str] = ..., url: _Optional[str] = ..., endpoint_id: _Optional[str] = ..., env: _Optional[_Mapping[str, str]] = ..., header: _Optional[_Mapping[str, HeaderFields]] = ..., runtime: _Optional[str] = ..., deployment_id: _Optional[str] = ..., id: _Optional[str] = ...) -> None: ...

class HeaderFields(_message.Message):
    __slots__ = ("fields",)
    FIELDS_FIELD_NUMBER: _ClassVar[int]
    fields: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, fields: _Optional[_Iterable[str]] = ...) -> None: ...

class HTTPResponse(_message.Message):
    __slots__ = ("body", "code", "request_id", "header")
    class HeaderEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: HeaderFields
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[HeaderFields, _Mapping]] = ...) -> None: ...
    BODY_FIELD_NUMBER: _ClassVar[int]
    CODE_FIELD_NUMBER: _ClassVar[int]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    HEADER_FIELD_NUMBER: _ClassVar[int]
    body: bytes
    code: int
    request_id: str
    header: _containers.MessageMap[str, HeaderFields]
    def __init__(self, body: _Optional[bytes] = ..., code: _Optional[int] = ..., request_id: _Optional[str] = ..., header: _Optional[_Mapping[str, HeaderFields]] = ...) -> None: ...
