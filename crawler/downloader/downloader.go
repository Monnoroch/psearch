package downloader

import (
	"io/ioutil"
	"net/http"
	"psearch/util/errors"
	"psearch/util/log"
)

type Downloader struct{}

func (self *Downloader) Download(url string) (string, error) {
	log.Printf("Downloader.Download(%s)\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.NewErr(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.NewErr(err)
	}

	log.Printf("Downloader.Download(%s) OK!\n", url)
	return string(body), nil
}

func (self *Downloader) DownloadAll(urls []string) ([]string, error) {
	log.Printf("Downloader.DownloadAll(%#v)!\n", urls)
	res := make([]string, len(urls))
	for i, url := range urls {
		b, err := self.Download(url)
		if err != nil {
			return nil, err
		}

		res[i] = b
	}
	log.Printf("Downloader.DownloadAll(%#v) OK!\n", urls)
	return res, nil
}

type DownloaderServer struct {
	Downloader *Downloader
}

func (self *DownloaderServer) Download(args *Args, result *string) error {
	r, err := self.Downloader.Download(args.Url)
	if err != nil {
		log.Errorln(err, args)
		return err
	}

	*result = r
	return nil
}

func (self *DownloaderServer) DownloadAll(args *ArgsAll, result *[]string) error {
	r, err := self.Downloader.DownloadAll(args.Urls)
	if err != nil {
		log.Errorln(err, args)
		return err
	}

	*result = r
	return nil
}
