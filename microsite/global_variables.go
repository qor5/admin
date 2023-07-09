package microsite

var PackageAndPreviewPrepath = "microsite"
var MaximumNumberOfFilesUploadedAtTheSameTime = 10
var putSemaphore = make(chan struct{}, MaximumNumberOfFilesUploadedAtTheSameTime)
