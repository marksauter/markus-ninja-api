package loader

// import (
//   "context"
//   "fmt"
//   "sync"
//
//   "github.com/graph-gophers/dataloader"
//   "github.com/marksauter/markus-ninja-api/pkg/data"
// )
//
// func NewEmailLoader(svc *data.EmailService) *EmailLoader {
//   return &EmailLoader{
//     svc:      svc,
//     batchGet: createLoader(newBatchGetEmailFn(svc.GetByPK)),
//   }
// }
//
// type EmailLoader struct {
//   svc *data.EmailService
//
//   batchGet *dataloader.Loader
// }
//
// func (r *EmailLoader) Clear(id string) {
//   ctx := context.Background()
//   r.batchGet.Clear(ctx, dataloader.StringKey(id))
// }
//
// func (r *EmailLoader) ClearAll() {
//   r.batchGet.ClearAll()
// }
//
// func (r *EmailLoader) Get(id string) (*data.Email, error) {
//   ctx := context.Background()
//   emailData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
//   if err != nil {
//     return nil, err
//   }
//   email, ok := emailData.(*data.Email)
//   if !ok {
//     return nil, fmt.Errorf("wrong type")
//   }
//
//   return email, nil
// }
//
// func newBatchGetEmailFn(
//   getter func(string) (*data.Email, error),
// ) dataloader.BatchFunc {
//   return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
//     var (
//       n       = len(keys)
//       results = make([]*dataloader.Result, n)
//       wg      sync.WaitGroup
//     )
//
//     wg.Add(n)
//
//     for i, key := range keys {
//       go func(i int, key dataloader.Key) {
//         defer wg.Done()
//         email, err := getter(key.String())
//         results[i] = &dataloader.Result{Data: email, Error: err}
//       }(i, key)
//     }
//
//     wg.Wait()
//
//     return results
//   }
// }
