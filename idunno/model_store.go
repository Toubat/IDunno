package main

import (
	"mp4/utils"
	"strings"
)

type ModelStore map[utils.ModelType]utils.DatasetType

func NewModelStore() *ModelStore {
	return &ModelStore{}
}

func (ms ModelStore) AddModel(modelType utils.ModelType, dataset utils.DatasetType) {
	ms[modelType] = utils.DatasetType(dataset)
}

func (ms ModelStore) Contains(modelType utils.ModelType) bool {
	_, ok := ms[modelType]
	return ok
}

func (ms ModelStore) GetDataset(modelType utils.ModelType) utils.DatasetType {
	return ms[modelType]
}

func IsValidModelType(modelType string) bool {
	switch utils.ModelType(modelType) {
	case utils.ResNet50, utils.Albert:
		return true
	default:
		return false
	}
}

func IsValidDatasetType(datasetType string) bool {
	switch utils.DatasetType(datasetType) {
	case utils.Imagenet, utils.Emotion:
		return true
	default:
		return false
	}
}

func ContainFilenames(dataset utils.DatasetType) bool {
	switch dataset {
	case utils.Imagenet:
		return true
	case utils.Emotion:
		return false
	}
	return false
}

func ValidateDatasetInput(dataset utils.DatasetType, input string) bool {
	switch dataset {
	case utils.Imagenet:
		return strings.HasSuffix(input, ".JPEG")
	case utils.Emotion:
		lst := strings.Split(input, ";")
		return len(lst) == 2
	}
	return false
}

func ValidateDatasetInputs(dataset utils.DatasetType, inputs []string) []string {
	if inputs == nil {
		return nil
	}

	validatedInputs := make([]string, 0)
	for _, input := range inputs {
		if !ValidateDatasetInput(dataset, input) {
			continue
		}
		validatedInputs = append(validatedInputs, input)
	}

	return validatedInputs
}
