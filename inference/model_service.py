from typing import List, Any, Tuple
from collections import OrderedDict

class IDunnoModelService:
    def __init__(self):
        pass

    def load_pretrained_model(self) -> Any:
        raise NotImplementedError("load_pretrained_model() not implemented")

    def parse_input(self, raw: Any) -> Any:
        raise NotImplementedError("parse_input() not implemented")

    def forward(self, inputs: List[Any]) -> List[Any]:
        raise NotImplementedError("forward() not implemented")

    def get_metrics(self, output: Any, input: Any) -> Any:
        raise NotImplementedError("get_metrics() not implemented")

    def inference(self, raw_inputs: List[Any]) -> Tuple[List[Any], List[Any]]:
        inputs = [self.parse_input(raw_input) for raw_input in raw_inputs]
        outputs = self.forward(inputs)

        metrics = [self.get_metrics(output, label) for output, label in zip(outputs, raw_inputs)]
        return (outputs, metrics)