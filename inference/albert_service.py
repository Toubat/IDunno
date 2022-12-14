import torch
from typing import List, Any
from model_service import IDunnoModelService
from transformers import pipeline, Pipeline

class AlbertModelService(IDunnoModelService):
    def __init__(self):
        super().__init__()

        self.model = self.load_pretrained_model()

    def load_pretrained_model(self) -> Pipeline:
        """
        Load the pretrained model

        Returns:
        """
        return pipeline("text-classification", model='bhadresh-savani/albert-base-v2-emotion')

    def parse_input(self, raw: str) -> str:
        """
        Parse the filename into an image.

        Args:
            raw (str): input text

        Returns:
            str: The parsed input.
        """
        raw = raw.split(';')
        return raw[0]

    def forward(self, inputs: List[str]) -> List[int]:
        """
        Forward pass of the model.

        Args:
            inputs (List[str]): The list of sentences.

        Returns:
            List[int]: The list of predictions.
        """
        return [res['label'] for res in self.model(inputs)]

    def get_metrics(self, output: int, label: str) -> int:
        """
        Get the metrics for the model.

        Args:
            output (int): The output of the model.
            label (str): The label of the text.

        Returns:
            int: The accuracy of the model on this input.
        """
        label = label.split(';')[1].strip()
        return 1 if output == label else 0
