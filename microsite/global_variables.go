package microsite

var PackageAndPreviewPrepath = "microsite"
var MaximumNumberOfFilesUploadedAtTheSameTime = 10
var putSemaphore = make(chan struct{}, MaximumNumberOfFilesUploadedAtTheSameTime)

var MaximumNumberOfFilesCopiedAtTheSameTime = 10
var copySemaphore = make(chan struct{}, MaximumNumberOfFilesCopiedAtTheSameTime)
