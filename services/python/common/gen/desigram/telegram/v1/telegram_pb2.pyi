from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class SendPhotoRequest(_message.Message):
    __slots__ = ["caption", "chat_id", "photo"]
    CAPTION_FIELD_NUMBER: _ClassVar[int]
    CHAT_ID_FIELD_NUMBER: _ClassVar[int]
    PHOTO_FIELD_NUMBER: _ClassVar[int]
    caption: str
    chat_id: int
    photo: bytes
    def __init__(self, chat_id: _Optional[int] = ..., photo: _Optional[bytes] = ..., caption: _Optional[str] = ...) -> None: ...

class SendPhotoResponse(_message.Message):
    __slots__ = ["message_id"]
    MESSAGE_ID_FIELD_NUMBER: _ClassVar[int]
    message_id: int
    def __init__(self, message_id: _Optional[int] = ...) -> None: ...
