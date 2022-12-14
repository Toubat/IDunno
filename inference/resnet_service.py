import torch
from PIL import Image
from transformers import AutoFeatureExtractor, ResNetForImageClassification

from typing import List, Any
from collections import OrderedDict
from model_service import IDunnoModelService
import json

class Resnet50ModelService(IDunnoModelService):

    CLASSES = json.load(open("../inference/resnet.json"))

    def __init__(self):
        super().__init__()

        extractor, resnet = self.load_pretrained_model()
        self.extractor = extractor
        self.resnet = resnet

    def load_pretrained_model(self) -> Any:
        """
        Load the pretrained model

        Returns:
            extractor: The feature extractor
            resnet: The resnet model
        """
        return (
            AutoFeatureExtractor.from_pretrained("microsoft/resnet-50"),
            ResNetForImageClassification.from_pretrained("microsoft/resnet-50"),
        )

    def parse_input(self, raw: str) -> Image.Image:
        """
        Parse the filename into an image.

        Args:
            raw (str): The filename of the image.

        Returns:
            Image.Image: The image.
        """
        image = Image.open(raw)

        # if not 224x224x3, extend to 3 channels
        if image.mode != "RGB":
            image = image.convert("RGB")

        return image

    def forward(self, inputs: List[Image.Image]) -> List[int]:
        """
        Forward pass of the model.

        Args:
            inputs (List[Image.Image]): The list of images.

        Returns:
            List[int]: The list of predictions.
        """
        inputs = self.extractor(inputs, return_tensors="pt")

        with torch.no_grad():
            logits = self.resnet(**inputs).logits

        predicted_class_idx = logits.argmax(-1)
        return [self.resnet.config.id2label[id] for id in predicted_class_idx.tolist()]

    def get_metrics(self, output: int, input: str) -> int:
        """
        Get the metrics for the model.

        Args:
            output (int): The output of the model.
            label (str): The label of the image.

        Returns:
            int: The accuracy of the model on this input.
        """
        label = input.split('_')[-1].split('.')[0]

        return 1 if output == self.CLASSES[label] else 0