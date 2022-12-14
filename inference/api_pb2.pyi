from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

Ack: MessageType
Alive: Status
Available: BatchStatus
Completed: BatchStatus
DESCRIPTOR: _descriptor.FileDescriptor
ERROR: ResponseStatus
InProgress: BatchStatus
Join: MessageType
Leave: MessageType
Leaved: Status
NOT_CONVERGED: ResponseStatus
NOT_FOUND: ResponseStatus
OK: ResponseStatus
Ping: MessageType
Timeout: Status

class AckMessage(_message.Message):
    __slots__ = ["received"]
    RECEIVED_FIELD_NUMBER: _ClassVar[int]
    received: str
    def __init__(self, received: _Optional[str] = ...) -> None: ...

class BackupRequest(_message.Message):
    __slots__ = ["backup"]
    BACKUP_FIELD_NUMBER: _ClassVar[int]
    backup: CoordinatorBackup
    def __init__(self, backup: _Optional[_Union[CoordinatorBackup, _Mapping]] = ...) -> None: ...

class BackupResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class BatchInput(_message.Message):
    __slots__ = ["batchId", "inputs"]
    BATCHID_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    batchId: int
    inputs: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, batchId: _Optional[int] = ..., inputs: _Optional[_Iterable[str]] = ...) -> None: ...

class BatchOutput(_message.Message):
    __slots__ = ["batchId", "metric", "results"]
    BATCHID_FIELD_NUMBER: _ClassVar[int]
    METRIC_FIELD_NUMBER: _ClassVar[int]
    RESULTS_FIELD_NUMBER: _ClassVar[int]
    batchId: int
    metric: float
    results: _containers.RepeatedCompositeFieldContainer[EvalResult]
    def __init__(self, batchId: _Optional[int] = ..., results: _Optional[_Iterable[_Union[EvalResult, _Mapping]]] = ..., metric: _Optional[float] = ...) -> None: ...

class BatchState(_message.Message):
    __slots__ = ["batchInput", "batchOutput", "queryTime", "receiveTime", "status"]
    BATCHINPUT_FIELD_NUMBER: _ClassVar[int]
    BATCHOUTPUT_FIELD_NUMBER: _ClassVar[int]
    QUERYTIME_FIELD_NUMBER: _ClassVar[int]
    RECEIVETIME_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    batchInput: BatchInput
    batchOutput: BatchOutput
    queryTime: _timestamp_pb2.Timestamp
    receiveTime: _timestamp_pb2.Timestamp
    status: BatchStatus
    def __init__(self, status: _Optional[_Union[BatchStatus, str]] = ..., batchInput: _Optional[_Union[BatchInput, _Mapping]] = ..., batchOutput: _Optional[_Union[BatchOutput, _Mapping]] = ..., queryTime: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., receiveTime: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class BulkLookupRequest(_message.Message):
    __slots__ = ["filenames", "seq"]
    FILENAMES_FIELD_NUMBER: _ClassVar[int]
    SEQ_FIELD_NUMBER: _ClassVar[int]
    filenames: _containers.RepeatedScalarFieldContainer[str]
    seq: Sequence
    def __init__(self, filenames: _Optional[_Iterable[str]] = ..., seq: _Optional[_Union[Sequence, _Mapping]] = ...) -> None: ...

class BulkLookupResponse(_message.Message):
    __slots__ = ["ip", "missingFiles", "port"]
    IP_FIELD_NUMBER: _ClassVar[int]
    MISSINGFILES_FIELD_NUMBER: _ClassVar[int]
    PORT_FIELD_NUMBER: _ClassVar[int]
    ip: str
    missingFiles: _containers.RepeatedScalarFieldContainer[str]
    port: int
    def __init__(self, ip: _Optional[str] = ..., port: _Optional[int] = ..., missingFiles: _Optional[_Iterable[str]] = ...) -> None: ...

class CoordinatorBackup(_message.Message):
    __slots__ = ["activeJobs", "completedJobs", "modelStore", "pendingJobs"]
    class ModelStoreEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    ACTIVEJOBS_FIELD_NUMBER: _ClassVar[int]
    COMPLETEDJOBS_FIELD_NUMBER: _ClassVar[int]
    MODELSTORE_FIELD_NUMBER: _ClassVar[int]
    PENDINGJOBS_FIELD_NUMBER: _ClassVar[int]
    activeJobs: _containers.RepeatedCompositeFieldContainer[Job]
    completedJobs: _containers.RepeatedCompositeFieldContainer[Job]
    modelStore: _containers.ScalarMap[str, str]
    pendingJobs: _containers.RepeatedCompositeFieldContainer[Job]
    def __init__(self, modelStore: _Optional[_Mapping[str, str]] = ..., activeJobs: _Optional[_Iterable[_Union[Job, _Mapping]]] = ..., completedJobs: _Optional[_Iterable[_Union[Job, _Mapping]]] = ..., pendingJobs: _Optional[_Iterable[_Union[Job, _Mapping]]] = ...) -> None: ...

class DeleteRequest(_message.Message):
    __slots__ = ["filename", "seq"]
    FILENAME_FIELD_NUMBER: _ClassVar[int]
    SEQ_FIELD_NUMBER: _ClassVar[int]
    filename: str
    seq: Sequence
    def __init__(self, filename: _Optional[str] = ..., seq: _Optional[_Union[Sequence, _Mapping]] = ...) -> None: ...

class DeleteResponse(_message.Message):
    __slots__ = ["status"]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: ResponseStatus
    def __init__(self, status: _Optional[_Union[ResponseStatus, str]] = ...) -> None: ...

class EvalResult(_message.Message):
    __slots__ = ["input", "output"]
    INPUT_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    input: str
    output: str
    def __init__(self, input: _Optional[str] = ..., output: _Optional[str] = ...) -> None: ...

class EvaluateRequest(_message.Message):
    __slots__ = ["inputs"]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    inputs: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, inputs: _Optional[_Iterable[str]] = ...) -> None: ...

class EvaluateResponse(_message.Message):
    __slots__ = ["metric", "results", "status"]
    METRIC_FIELD_NUMBER: _ClassVar[int]
    RESULTS_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    metric: float
    results: _containers.RepeatedCompositeFieldContainer[EvalResult]
    status: ResponseStatus
    def __init__(self, results: _Optional[_Iterable[_Union[EvalResult, _Mapping]]] = ..., metric: _Optional[float] = ..., status: _Optional[_Union[ResponseStatus, str]] = ...) -> None: ...

class FetchSequenceRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class FetchSequenceResponse(_message.Message):
    __slots__ = ["seq", "status"]
    SEQ_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    seq: Sequence
    status: ResponseStatus
    def __init__(self, status: _Optional[_Union[ResponseStatus, str]] = ..., seq: _Optional[_Union[Sequence, _Mapping]] = ...) -> None: ...

class FinishInferenceRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class FinishInferenceResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GreetRequest(_message.Message):
    __slots__ = ["name"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    name: str
    def __init__(self, name: _Optional[str] = ...) -> None: ...

class GreetResponse(_message.Message):
    __slots__ = ["message"]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    message: str
    def __init__(self, message: _Optional[str] = ...) -> None: ...

class HeartbeatRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class HeartbeatResponse(_message.Message):
    __slots__ = ["status"]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: ResponseStatus
    def __init__(self, status: _Optional[_Union[ResponseStatus, str]] = ...) -> None: ...

class IDunnoStatusRequest(_message.Message):
    __slots__ = ["payload", "which"]
    PAYLOAD_FIELD_NUMBER: _ClassVar[int]
    WHICH_FIELD_NUMBER: _ClassVar[int]
    payload: str
    which: str
    def __init__(self, which: _Optional[str] = ..., payload: _Optional[str] = ...) -> None: ...

class IDunnoStatusResponse(_message.Message):
    __slots__ = ["message"]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    message: str
    def __init__(self, message: _Optional[str] = ...) -> None: ...

class InferenceRequest(_message.Message):
    __slots__ = ["inferenceTask", "jobId"]
    INFERENCETASK_FIELD_NUMBER: _ClassVar[int]
    JOBID_FIELD_NUMBER: _ClassVar[int]
    inferenceTask: InferenceTask
    jobId: str
    def __init__(self, inferenceTask: _Optional[_Union[InferenceTask, _Mapping]] = ..., jobId: _Optional[str] = ...) -> None: ...

class InferenceResponse(_message.Message):
    __slots__ = ["status"]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: ResponseStatus
    def __init__(self, status: _Optional[_Union[ResponseStatus, str]] = ...) -> None: ...

class InferenceTask(_message.Message):
    __slots__ = ["batchSize", "model"]
    BATCHSIZE_FIELD_NUMBER: _ClassVar[int]
    MODEL_FIELD_NUMBER: _ClassVar[int]
    batchSize: int
    model: str
    def __init__(self, model: _Optional[str] = ..., batchSize: _Optional[int] = ...) -> None: ...

class Job(_message.Message):
    __slots__ = ["batchSize", "batchStates", "completedQueries", "dataset", "finishTime", "id", "modelType", "queryProcessTimes", "queryRates", "startTime", "totalQueries"]
    BATCHSIZE_FIELD_NUMBER: _ClassVar[int]
    BATCHSTATES_FIELD_NUMBER: _ClassVar[int]
    COMPLETEDQUERIES_FIELD_NUMBER: _ClassVar[int]
    DATASET_FIELD_NUMBER: _ClassVar[int]
    FINISHTIME_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    MODELTYPE_FIELD_NUMBER: _ClassVar[int]
    QUERYPROCESSTIMES_FIELD_NUMBER: _ClassVar[int]
    QUERYRATES_FIELD_NUMBER: _ClassVar[int]
    STARTTIME_FIELD_NUMBER: _ClassVar[int]
    TOTALQUERIES_FIELD_NUMBER: _ClassVar[int]
    batchSize: int
    batchStates: _containers.RepeatedCompositeFieldContainer[BatchState]
    completedQueries: int
    dataset: str
    finishTime: _timestamp_pb2.Timestamp
    id: str
    modelType: str
    queryProcessTimes: _containers.RepeatedScalarFieldContainer[float]
    queryRates: _containers.RepeatedScalarFieldContainer[float]
    startTime: _timestamp_pb2.Timestamp
    totalQueries: int
    def __init__(self, id: _Optional[str] = ..., modelType: _Optional[str] = ..., dataset: _Optional[str] = ..., batchSize: _Optional[int] = ..., startTime: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., finishTime: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., totalQueries: _Optional[int] = ..., completedQueries: _Optional[int] = ..., batchStates: _Optional[_Iterable[_Union[BatchState, _Mapping]]] = ..., queryRates: _Optional[_Iterable[float]] = ..., queryProcessTimes: _Optional[_Iterable[float]] = ...) -> None: ...

class JoinMessage(_message.Message):
    __slots__ = ["process"]
    PROCESS_FIELD_NUMBER: _ClassVar[int]
    process: Process
    def __init__(self, process: _Optional[_Union[Process, _Mapping]] = ...) -> None: ...

class LeaveMessage(_message.Message):
    __slots__ = ["process"]
    PROCESS_FIELD_NUMBER: _ClassVar[int]
    process: Process
    def __init__(self, process: _Optional[_Union[Process, _Mapping]] = ...) -> None: ...

class LookupLeaderRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class LookupLeaderResponse(_message.Message):
    __slots__ = ["address"]
    ADDRESS_FIELD_NUMBER: _ClassVar[int]
    address: str
    def __init__(self, address: _Optional[str] = ...) -> None: ...

class LookupRequest(_message.Message):
    __slots__ = ["filename", "seq"]
    FILENAME_FIELD_NUMBER: _ClassVar[int]
    SEQ_FIELD_NUMBER: _ClassVar[int]
    filename: str
    seq: Sequence
    def __init__(self, filename: _Optional[str] = ..., seq: _Optional[_Union[Sequence, _Mapping]] = ...) -> None: ...

class LookupResponse(_message.Message):
    __slots__ = ["ip", "port", "status"]
    IP_FIELD_NUMBER: _ClassVar[int]
    PORT_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ip: str
    port: int
    status: ResponseStatus
    def __init__(self, ip: _Optional[str] = ..., port: _Optional[int] = ..., status: _Optional[_Union[ResponseStatus, str]] = ...) -> None: ...

class Metadata(_message.Message):
    __slots__ = ["ack", "join", "leave", "ping", "type"]
    ACK_FIELD_NUMBER: _ClassVar[int]
    JOIN_FIELD_NUMBER: _ClassVar[int]
    LEAVE_FIELD_NUMBER: _ClassVar[int]
    PING_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    ack: AckMessage
    join: JoinMessage
    leave: LeaveMessage
    ping: PingMessage
    type: MessageType
    def __init__(self, type: _Optional[_Union[MessageType, str]] = ..., ping: _Optional[_Union[PingMessage, _Mapping]] = ..., ack: _Optional[_Union[AckMessage, _Mapping]] = ..., join: _Optional[_Union[JoinMessage, _Mapping]] = ..., leave: _Optional[_Union[LeaveMessage, _Mapping]] = ...) -> None: ...

class PingMessage(_message.Message):
    __slots__ = ["processes"]
    PROCESSES_FIELD_NUMBER: _ClassVar[int]
    processes: _containers.RepeatedCompositeFieldContainer[Process]
    def __init__(self, processes: _Optional[_Iterable[_Union[Process, _Mapping]]] = ...) -> None: ...

class Process(_message.Message):
    __slots__ = ["ip", "joinTime", "lastUpdateTime", "port", "status"]
    IP_FIELD_NUMBER: _ClassVar[int]
    JOINTIME_FIELD_NUMBER: _ClassVar[int]
    LASTUPDATETIME_FIELD_NUMBER: _ClassVar[int]
    PORT_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ip: str
    joinTime: _timestamp_pb2.Timestamp
    lastUpdateTime: _timestamp_pb2.Timestamp
    port: int
    status: Status
    def __init__(self, ip: _Optional[str] = ..., port: _Optional[int] = ..., joinTime: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., lastUpdateTime: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., status: _Optional[_Union[Status, str]] = ...) -> None: ...

class QueryDataRequest(_message.Message):
    __slots__ = ["batchOutput", "jobId", "worker"]
    BATCHOUTPUT_FIELD_NUMBER: _ClassVar[int]
    JOBID_FIELD_NUMBER: _ClassVar[int]
    WORKER_FIELD_NUMBER: _ClassVar[int]
    batchOutput: BatchOutput
    jobId: str
    worker: Process
    def __init__(self, jobId: _Optional[str] = ..., worker: _Optional[_Union[Process, _Mapping]] = ..., batchOutput: _Optional[_Union[BatchOutput, _Mapping]] = ...) -> None: ...

class QueryDataResponse(_message.Message):
    __slots__ = ["batchInput", "isFilename"]
    BATCHINPUT_FIELD_NUMBER: _ClassVar[int]
    ISFILENAME_FIELD_NUMBER: _ClassVar[int]
    batchInput: BatchInput
    isFilename: bool
    def __init__(self, batchInput: _Optional[_Union[BatchInput, _Mapping]] = ..., isFilename: bool = ...) -> None: ...

class ReadRequest(_message.Message):
    __slots__ = ["filename", "localFilename", "seq", "version"]
    FILENAME_FIELD_NUMBER: _ClassVar[int]
    LOCALFILENAME_FIELD_NUMBER: _ClassVar[int]
    SEQ_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    filename: str
    localFilename: str
    seq: Sequence
    version: int
    def __init__(self, filename: _Optional[str] = ..., version: _Optional[int] = ..., localFilename: _Optional[str] = ..., seq: _Optional[_Union[Sequence, _Mapping]] = ...) -> None: ...

class ReadResponse(_message.Message):
    __slots__ = ["data", "seq", "status"]
    DATA_FIELD_NUMBER: _ClassVar[int]
    SEQ_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    data: bytes
    seq: Sequence
    status: ResponseStatus
    def __init__(self, data: _Optional[bytes] = ..., status: _Optional[_Union[ResponseStatus, str]] = ..., seq: _Optional[_Union[Sequence, _Mapping]] = ...) -> None: ...

class Sequence(_message.Message):
    __slots__ = ["count", "time"]
    COUNT_FIELD_NUMBER: _ClassVar[int]
    TIME_FIELD_NUMBER: _ClassVar[int]
    count: int
    time: _timestamp_pb2.Timestamp
    def __init__(self, time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., count: _Optional[int] = ...) -> None: ...

class ServeModelRequest(_message.Message):
    __slots__ = ["model"]
    MODEL_FIELD_NUMBER: _ClassVar[int]
    model: str
    def __init__(self, model: _Optional[str] = ...) -> None: ...

class ServeModelResponse(_message.Message):
    __slots__ = ["status"]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: ResponseStatus
    def __init__(self, status: _Optional[_Union[ResponseStatus, str]] = ...) -> None: ...

class TrainRequest(_message.Message):
    __slots__ = ["trainTask"]
    TRAINTASK_FIELD_NUMBER: _ClassVar[int]
    trainTask: TrainTask
    def __init__(self, trainTask: _Optional[_Union[TrainTask, _Mapping]] = ...) -> None: ...

class TrainResponse(_message.Message):
    __slots__ = ["status"]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: ResponseStatus
    def __init__(self, status: _Optional[_Union[ResponseStatus, str]] = ...) -> None: ...

class TrainTask(_message.Message):
    __slots__ = ["dataset", "model"]
    DATASET_FIELD_NUMBER: _ClassVar[int]
    MODEL_FIELD_NUMBER: _ClassVar[int]
    dataset: str
    model: str
    def __init__(self, model: _Optional[str] = ..., dataset: _Optional[str] = ...) -> None: ...

class UpdateLeaderRequest(_message.Message):
    __slots__ = ["leader"]
    LEADER_FIELD_NUMBER: _ClassVar[int]
    leader: Process
    def __init__(self, leader: _Optional[_Union[Process, _Mapping]] = ...) -> None: ...

class UpdateLeaderResponse(_message.Message):
    __slots__ = ["status"]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: ResponseStatus
    def __init__(self, status: _Optional[_Union[ResponseStatus, str]] = ...) -> None: ...

class WriteId(_message.Message):
    __slots__ = ["createTime", "ip", "port"]
    CREATETIME_FIELD_NUMBER: _ClassVar[int]
    IP_FIELD_NUMBER: _ClassVar[int]
    PORT_FIELD_NUMBER: _ClassVar[int]
    createTime: _timestamp_pb2.Timestamp
    ip: str
    port: int
    def __init__(self, ip: _Optional[str] = ..., port: _Optional[int] = ..., createTime: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class WriteRequest(_message.Message):
    __slots__ = ["data", "filename", "seq", "writeId"]
    DATA_FIELD_NUMBER: _ClassVar[int]
    FILENAME_FIELD_NUMBER: _ClassVar[int]
    SEQ_FIELD_NUMBER: _ClassVar[int]
    WRITEID_FIELD_NUMBER: _ClassVar[int]
    data: bytes
    filename: str
    seq: Sequence
    writeId: WriteId
    def __init__(self, filename: _Optional[str] = ..., data: _Optional[bytes] = ..., writeId: _Optional[_Union[WriteId, _Mapping]] = ..., seq: _Optional[_Union[Sequence, _Mapping]] = ...) -> None: ...

class WriteResponse(_message.Message):
    __slots__ = ["status"]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: ResponseStatus
    def __init__(self, status: _Optional[_Union[ResponseStatus, str]] = ...) -> None: ...

class Status(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []

class MessageType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []

class ResponseStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []

class BatchStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
