//go:build windows

package persistence

import (
 "context"
 "errors"
 "io"
 "golang.org/x/sys/windows"
)

func reviewWriteCurrent(root string,raw []byte)(resultErr error){
 c,err:=nativeAcquire(context.Background(),root);if err!=nil{return err}
 defer func(){resultErr=errors.Join(resultErr,c.close())}()
 f,err:=winOpen(c.parent().handle(),"config.json",windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE,winShareAll,windows.FILE_OPEN,windows.FILE_NON_DIRECTORY_FILE)
 if err!=nil{return err};defer func(){resultErr=errors.Join(resultErr,f.close())}()
 n,err:=f.file.WriteAt(raw,0);if err!=nil{return err};if n!=len(raw){return io.ErrShortWrite};return nil
}
