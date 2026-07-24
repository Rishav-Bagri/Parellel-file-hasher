package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"
)

type result struct {
	path string
	hash string
}

func hashKarunga(path string , ans chan result){
	file,err:= os.Open(path)
	if err!=nil{
		log.Fatal("cant open file")
	}
	defer file.Close()
	
	hasher := sha256.New()	
	
	if _,err:= io.Copy(hasher,file) ; err!=nil{

		ans<- result{path, ""}
		return
	}

	hashBytes:= hasher.Sum(nil)

	ans<-result{
		path: path,
		hash:hex.EncodeToString(hashBytes),
	}

}

func worker(job chan string, ans chan result, done chan bool){
	for	a:= range job{
		hashKarunga( a, ans)
	}
	done<-true
}


func walking(rootPath string,job chan string, res chan result){
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // Handle access errors or permissions issues safely
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		job<- path
		
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking path: %v\n", err)
	}
	
	close (job)


}


func main() {
	rootPath := "C:/Users/aecr/Desktop/project"
	len:=0

	noOfWorker:=800

	job:= make(chan string)
	ans:= make(chan result)
	done:= make(chan bool)
	fmt.Println("start")
	startTimer:= time.Now()

	go walking(rootPath, job, ans)

	for range noOfWorker{
		go worker(job, ans,done)
	}

	res:= map[string]int{}

	go func(){
		
		for x := range ans {
			res[x.hash]++
		}
	}()
	for i:=0;i<noOfWorker;i++{
		<-done
	}

	close(ans)
	for k,v:=range res{
		len++
		fmt.Println(k," ",v)
	}
	fmt.Println(len)
	timeTaken:=time.Since(startTimer)

	fmt.Println("Time: ",timeTaken)
}
