package acoustid

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var mp3FileCount = 0

var (
	fingerPrintInputChan = make(chan FingerprintInput, 100)
	acousticIdInputChan  = make(chan Acousticidinput, 100)
	id3TagInputChan      = make(chan ID3TagInput, 100)
	processedChan        = make(chan bool, 100)
)

var numCores = 0

func init() {
	numCores = runtime.NumCPU()
}

func TagDirParallel(dir string) {
	for i := 0; i < numCores; i++ {
		go FingerprintWorker()
		go AcousticidWorker()
		go ID3Worker()
	}

	//Count the number of files to tag:
	filepath.Walk(dir, countMp3Files)

	//Send the files to be fingerprinted
	filepath.Walk(dir, tagFileParallel)

	//Ensure all files have been processed
	for i := 0; i < mp3FileCount; i++ {
		<-processedChan
	}
}

func countMp3Files(file string, info os.FileInfo, err error) error {
	if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".mp3") {
		mp3FileCount++
	}
	return nil
}

func tagFileParallel(file string, info os.FileInfo, err error) error {
	if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".mp3") {
		fingerPrintInputChan <- FingerprintInput{file, info}
	}
	return nil
}

func FingerprintWorker() {
	for fingerprintInput := range fingerPrintInputChan {
		acousticIdInputChan <- Acousticidinput{fingerprintInput.file, fingerprintInput.info, NewFingerprint(fingerprintInput.file)}
	}
}

func AcousticidWorker() {
	for acousticidInput := range acousticIdInputChan {
		id3TagInputChan <- ID3TagInput{acousticidInput.file, acousticidInput.info, MakeAcoustIDRequest(acousticidInput.fingerprint)}
	}
}

func ID3Worker() {
	for id3TagInput := range id3TagInputChan {
		SetID3(id3TagInput.resp, id3TagInput.file, id3TagInput.info)
		processedChan <- true
	}
}

//Structs to pass around the channels
type FingerprintInput struct {
	file string
	info os.FileInfo
}

type Acousticidinput struct {
	file        string
	info        os.FileInfo
	fingerprint Fingerprint
}

type ID3TagInput struct {
	file string
	info os.FileInfo
	resp AcoustIDResponse
}
