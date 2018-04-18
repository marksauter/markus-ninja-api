package service

//
// import (
//   "database/sql"
//   "errors"
//   "fmt"
//
//   "github.com/marksauter/markus-ninja-api/pkg/attr"
//   "github.com/marksauter/markus-ninja-api/pkg/model"
//   "github.com/marksauter/markus-ninja-api/pkg/mydb"
//   "github.com/marksauter/markus-ninja-api/pkg/mylog"
//   "github.com/marksauter/markus-ninja-api/pkg/passwd"
//   "github.com/marksauter/markus-ninja-api/pkg/util"
// )
//
// const (
//   defaultListFetchSize = 10
// )
//
// func NewStudyService(db *mydb.DB, roleSvc *RoleService) *StudyService {
//   return &StudyService{db: db, roleSvc: roleSvc}
// }
//
// type StudyService struct {
//   db *mydb.DB
// }
//
// func (s *StudyService) Get(id string) (*model.Study, error) {
//   study := new(model.Study)
//
//   studySQL := `SELECT * FROM studys WHERE id = $1`
//   row := s.db.QueryRowx(studySQL, id)
//   err := row.StructScan(study)
//   if err != nil {
//     switch err {
//     case sql.ErrNoRows:
//       return study, nil
//     default:
//       mylog.Log.WithField("error", err).Errorf("Get(%v)", id)
//       return nil, err
//     }
//   }
//
//   roles, err := s.roleSvc.GetByStudyId(study.ID)
//   if err != nil {
//     mylog.Log.WithField("error", err).Errorf("Get(%v)", id)
//     return nil, err
//   }
//   study.Roles = roles
//
//   mylog.Log.WithField("study", study).Debugf("Get(%v)", id)
//   return study, nil
// }
//
// func (s *StudyService) BatchGet(ids []string) ([]model.Study, error) {
//   studys := []model.Study{}
//
//   whereIn := "$1"
//   for i, _ := range ids[0:] {
//     whereIn = whereIn + fmt.Sprintf(", $%v", i+1)
//   }
//   batchGetSQL := fmt.Sprintf("SELECT * FROM studys WHERE id IN (%v)", whereIn)
//
//   err := s.db.Select(&studys, batchGetSQL, util.StringToInterface(ids)...)
//   if err != nil {
//     mylog.Log.WithField("error", err).Errorf("BatchGet(%v)", ids)
//     return nil, err
//   }
//
//   mylog.Log.WithField("studys", studys).Debugf("BatchGet(%v)", ids)
//   return studys, nil
// }
//
// func (s *StudyService) GetByLogin(login string) (*model.Study, error) {
//   study := new(model.Study)
//
//   studySQL := `SELECT * FROM studys WHERE login = $1`
//   row := s.db.QueryRowx(studySQL, login)
//   err := row.StructScan(study)
//   if err != nil {
//     mylog.Log.WithField("error", err).Errorf("GetByLogin(%v)", login)
//     return nil, err
//   }
//
//   roles, err := s.roleSvc.GetByStudyId(study.ID)
//   if err != nil {
//     mylog.Log.WithField("error", err).Errorf("GetByLogin(%v)", login)
//     return nil, err
//   }
//   study.Roles = roles
//
//   mylog.Log.WithField("study", study).Debugf("GetByLogin(%v)", login)
//   return study, nil
// }
//
// type CreateStudyInput struct {
//   Bio      string
//   Login    string
//   Password string
// }
//
// func (s *StudyService) Create(input *CreateStudyInput) (*model.Study, error) {
//   studyID := attr.NewId("Study")
//   password := passwd.New(input.Password)
//   if ok := password.CheckStrength(passwd.VeryWeak); !ok {
//     mylog.Log.WithField(
//       "error", "password failed strength check",
//     ).Errorf("Create(%+v)", input)
//     return new(model.Study), errors.New("Password too weak")
//   }
//   pwdHash, err := password.Hash()
//   if err != nil {
//     mylog.Log.WithField("error", err).Errorf("Create(%+v)", input)
//     return nil, err
//   }
//   study := model.Study{
//     Bio:      input.Bio,
//     ID:       studyID.String(),
//     Login:    input.Login,
//     Password: pwdHash,
//   }
//
//   studySQL := `INSERT INTO studys (id, login, password) VALUES (:id, :login, :password)`
//   _, err = s.db.NamedExec(studySQL, study)
//   if err != nil {
//     mylog.Log.WithField("error", err).Errorf("Create(%+v)", input)
//     return nil, err
//   }
//
//   mylog.Log.WithField("study", study).Debugf("Create(%+v)", input)
//   return &study, nil
// }
//
// func (s *StudyService) VerifyCredentials(studyCredentials *model.StudyCredentials) (*model.Study, error) {
//   study, err := s.GetByLogin(studyCredentials.Login)
//   if err != nil {
//     mylog.Log.WithField("error", err).Errorf("VerifyCredentials(%+v)", studyCredentials)
//     return nil, errors.New("unauthorized access")
//   }
//   password := passwd.New(studyCredentials.Password)
//   if match := password.CompareToHash([]byte(study.Password)); !match {
//     mylog.Log.WithField(
//       "error", "password doesn't match hash",
//     ).Errorf("VerifyCredentials(%+v)", studyCredentials)
//     return nil, errors.New("unauthorized access")
//   }
//
//   mylog.Log.WithField(
//     "study", study,
//   ).Debugf("VerifyCredentials(%+v)", studyCredentials)
//   return study, nil
// }
