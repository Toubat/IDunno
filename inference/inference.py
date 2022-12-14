from typing import Optional
from enum import Enum

from model_service import IDunnoModelService
from resnet_service import Resnet50ModelService
from albert_service import AlbertModelService
from api_pb2_grpc import InferenceService
from api_pb2 import OK, ERROR
from api_pb2 import TrainTask
from api_pb2 import (
    GreetRequest,
    GreetResponse,
    TrainRequest,
    TrainResponse,
    ServeModelRequest,
    ServeModelResponse,
    EvaluateRequest,
    EvaluateResponse,
    EvalResult,
)


class ModelType(Enum):
    Resnet50 = "resnet50"
    Albert = "albert"


def start_inference_service(model: str) -> IDunnoModelService:
    match model:
        case ModelType.Resnet50.value:
            print("Starting Resnet50 model service")
            return Resnet50ModelService()
        case ModelType.Albert.value:
            print("Starting Albert model service")
            return AlbertModelService()
        case _:
            print("Unknown model type")
            raise Exception("Unknown model type")


class InferenceServiceServer(InferenceService):
    def __init__(self):
        super().__init__()
        self.model_service: Optional[IDunnoModelService] = None

    def Greet(self, request: GreetRequest, context) -> GreetResponse:
        return GreetResponse(message="Hello, %s!" % request.name)

    def Train(self, request: TrainRequest, context) -> TrainResponse:
        try:
            start_inference_service(request.trainTask.model)
            print("Training model completed")
            return TrainResponse(status=OK)
        except Exception as e:
            print("Training model failed: " + str(e))
            return TrainResponse(status=ERROR)

    def ServeModel(self, request: ServeModelRequest, context) -> ServeModelResponse:
        try:
            self.model_service = start_inference_service(request.model)
            print("Serving model completed")
            return ServeModelResponse(status=OK)
        except Exception as e:
            print("Serving model failed " + str(e))
            return ServeModelResponse(status=ERROR)

    def Evaluate(self, request: EvaluateRequest, context) -> EvaluateResponse:
        inputs = list(request.inputs)
        print("Model runner received %d inputs" % len(inputs))

        try:
            outputs, metrics = self.model_service.inference(inputs)
            print("Model runner completed")
            results = [EvalResult(input=inp, output=str(out)) for inp, out in zip(inputs, outputs)]
            metric = float(sum(metrics))
            return EvaluateResponse(results=results, metric=metric, status=OK)
        except Exception as e:
            print("Evaluating model failed: " + str(e))
            return EvaluateResponse(results=[], metric=0, status=ERROR)