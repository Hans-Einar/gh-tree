//go:build linux

package persistence

import (
 "context"
 "os"
 "path/filepath"
 "testing"
 "github.com/Hans-Einar/gh-tree/internal/application/api"
 "golang.org/x/sys/unix"
)

func reviewWriteCurrent(root string, raw []byte)error{return os.WriteFile(filepath.Join(root,"config.json"),raw,0600)}

func TestReviewUnixStagingSubstitution(t *testing.T) {
 for _, present := range []bool{false,true} {
  t.Run(map[bool]string{false:"absent",true:"present"}[present],func(t *testing.T){
   root := physicalStoreTemp(t)
   target:=filepath.Join(root,"config.json")
   if present { if err:=os.WriteFile(target,[]byte(`{"schemaVersion":1}`),0600);err!=nil{t.Fatal(err)} }
   outside:=filepath.Join(physicalStoreTemp(t),"outside.json")
   sentinel:=[]byte(`{"schemaVersion":1,"stripPrefixes":["external"]}`)
   if err:=os.WriteFile(outside,sentinel,0600);err!=nil{t.Fatal(err)}
   s:=newTestStore(t,root)
   loaded,err:=s.LoadUserConfig(context.Background());if err!=nil{t.Fatal(err)}
   v,_:=loaded.Observation().Data().Version.Value()
   s.hook=func(stage string)error{
    if stage!="final-check"{return nil}
    names,err:=filepath.Glob(filepath.Join(root,"*.publication"));if err!=nil{return err}
    if len(names)!=1{t.Fatalf("publication names: %v",names)}
    if err:=os.Rename(names[0],names[0]+".retained-by-test");err!=nil{return err}
    return os.Symlink(outside,names[0])
   }
   result,err:=s.CommitUserConfig(context.Background(),userProposal(t,v,"proposal"))
   st,statErr:=os.Lstat(target);if statErr!=nil{t.Fatal(statErr)}
   after,readErr:=os.ReadFile(outside);if readErr!=nil||string(after)!=string(sentinel){t.Fatal("external sentinel changed",readErr)}
   t.Logf("result valid=%v outcome=%v known=%v target_symlink=%v error=%v",result.Valid(),result.Data().Outcome,result.Data().PublicationKnown,st.Mode()&os.ModeSymlink!=0,err)
   if result.Data().PublicationKnown && st.Mode()&os.ModeSymlink!=0 { t.Fatal("verified retained proposal replaced by external symlink at publication") }
  })
 }
}

func TestReviewUnixModePreservedAfterWriting(t *testing.T){
 root:=physicalStoreTemp(t)
 target:=filepath.Join(root,"config.json")
 if err:=os.WriteFile(target,[]byte(`{"schemaVersion":1}`),0600);err!=nil{t.Fatal(err)}
 if err:=unix.Chmod(target,04750);err!=nil{t.Fatal(err)}
 var before unix.Stat_t;if err:=unix.Stat(target,&before);err!=nil{t.Fatal(err)}
 if before.Mode&07777!=04750{t.Fatal("fixture did not establish requested mode")}
 s:=newTestStore(t,root)
 loaded,err:=s.LoadUserConfig(context.Background());if err!=nil{t.Fatal(err)}
 v,_:=loaded.Observation().Data().Version.Value()
 result,err:=s.CommitUserConfig(context.Background(),userProposal(t,v,"proposal"))
 var after unix.Stat_t;if err:=unix.Stat(target,&after);err!=nil{t.Fatal(err)}
 t.Logf("result valid=%v outcome=%v known=%v mode_before=%o mode_after=%o error=%v",result.Valid(),result.Data().Outcome,result.Data().PublicationKnown,before.Mode&07777,after.Mode&07777,err)
 if result.Data().PublicationKnown && before.Mode&07777!=after.Mode&07777 {t.Fatal("commit silently changed preserved mode after payload write")}
}

func TestReviewUnixRecoveryRefusalPreservesCurrent(t *testing.T){
 root:=physicalStoreTemp(t);s:=newTestStore(t,root)
 loaded,err:=s.LoadUserConfig(context.Background());if err!=nil{t.Fatal(err)}
 v,_:=loaded.Observation().Data().Version.Value()
 result,err:=s.CommitUserConfig(context.Background(),userProposal(t,v,"proposal"));assertCommitted(t,result,err)
 outside:=filepath.Join(physicalStoreTemp(t),"outside")
 if err:=os.WriteFile(outside,[]byte("retained sentinel"),0600);err!=nil{t.Fatal(err)}
 if err:=os.Symlink(outside,filepath.Join(root,recoveryPrefix("config.json")+"review-orphan"));err!=nil{t.Fatal(err)}
 after,err:=newTestStore(t,root).LoadUserConfig(context.Background())
 t.Logf("load valid=%v document=%v state=%v version=%v error=%v",after.Valid(),after.Document().Present(),after.Observation().Data().State,after.Observation().Data().Version.Present(),err)
 if err==nil{t.Fatal("unsupported recovery link was silently accepted")}
 if !after.Valid()||!after.Document().Present()||after.Observation().Data().State!=api.ValidCurrent||!after.Observation().Data().Version.Present(){t.Fatal("recovery refusal erased independent usable current observation")}
}
