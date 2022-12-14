package utils

type ModelType string

const (
	ResNet50 ModelType = "resnet50"
	Albert   ModelType = "albert"
)

var SupportedModelTypes = []ModelType{ResNet50, Albert}

type DatasetType string

const (
	Imagenet DatasetType = "imagenet"
	Emotion  DatasetType = "emotion.txt"
)

var SupportedDatasetTypes = []DatasetType{Imagenet, Emotion}
