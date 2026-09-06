package persistence

import (
 "context"
 "errors"
 "testing"
 "github.com/Hans-Einar/gh-tree/internal/application/api"
)

func TestReviewKnownCommitPreservesCanceledDelivery(t *testing.T) {
 root:=physicalStoreTemp(t);s:=newTestStore(t,root)
 loaded,err:=s.LoadUserConfig(context.Background());if err!=nil{t.Fatal(err)}
 v,_:=loaded.Observation().Data().Version.Value()
 ctx,cancel:=context.WithCancel(context.Background());defer cancel()
 delivery:=errors.New("review delivery error")
 s.hook=func(stage string)error{if stage=="outcome-delivery"{cancel();return delivery};return nil}
 result,err:=s.CommitUserConfig(ctx,userProposal(t,v,"proposal"))
 d:=result.Data()
 if !errors.Is(err,delivery)||!result.Valid()||!d.PublicationKnown||!d.CancellationAsked||len(d.Recovery)<2{t.Fatalf("known effect lost: %+v %v",d,err)}
 if d.Outcome!=api.Committed&&d.Outcome!=api.CommittedDurabilityUncertain{t.Fatal("commit relabeled")}
 t.Logf("known commit survives delivery error/cancel with %d recovery records",len(d.Recovery))
}

func TestReviewKnownCommitIndependentCurrent(t *testing.T) {
 root:=physicalStoreTemp(t);s:=newTestStore(t,root)
 loaded,err:=s.LoadUserConfig(context.Background());if err!=nil{t.Fatal(err)}
 v,_:=loaded.Observation().Data().Version.Value()
 s.hook=func(stage string)error{if stage=="directory-flush"{return reviewWriteCurrent(root,[]byte(`{"schemaVersion":1,"stripPrefixes":["external later edit"]}`))};return nil}
 result,err:=s.CommitUserConfig(context.Background(),userProposal(t,v,"proposal"))
 if err!=nil||!result.Valid()||!result.Data().PublicationKnown{t.Fatal("later edit erased known publication",err)}
 p,pok:=result.Data().ProposedVersion.Value();c,cok:=result.Data().CurrentVersion.Value()
 if !pok||!cok||p==c{t.Fatal("later current bytes conflated with proposal")}
 t.Log("known publication retains distinct proposed/current content versions")
}
