package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type DBConfig struct {
	DB_Host     string
	DB_Port     int64
	DB_User     string
	DB_Password string
	DB_Name     string
}

type Register struct {
	GroupID              int8   `json:"group_id"`
	Name                 string `json:"name"`
	CompanyName          string `json:"company_name"`
	Gender               int8   `json:"gender"`
	BirthDate            string `json:"birth_date"`
	JobID                int8   `json:"job_id"`
	Address              string `json:"address"`
	ProvinceID           int    `json:"province_id"`
	CityID               int    `json:"city_id"`
	PhoneNumber          string `json:"phone_number"`
	IdentityType         int8   `json:"identity_type"`
	IdentityFile         string `json:"identity_file"`
	NpwpFile             string `json:"npwp_file"`
	SiupFile             string `json:"siup_file"`
	Email                string `json:"email"`
	Password             []byte `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

type Login struct {
	Email    string `json:"email"`
	Password []byte `json:"password"`
}

type VerificationAccount struct {
	VerificationCode string `json:"verification_code"`
}

var err error

func DBConnection() *sql.DB {
	var dbconfig DBConfig

	_, err := toml.DecodeFile(".env.toml", &dbconfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbconfig.DB_Host, dbconfig.DB_Port, dbconfig.DB_User, dbconfig.DB_Password, dbconfig.DB_Name)

	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		log.Fatal(err.Error())
	}

	return db
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", IndexPage).Methods("GET")
	r.HandleFunc("/auth/register", RegisterPage).Methods("POST")
	r.HandleFunc("/auth/verification-account", VerificationAccountPage).Methods("POST")
	r.HandleFunc("/auth/login", LoginPage).Methods("POST")

	fmt.Print("Starting web server at http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// Default Router
func IndexPage(w http.ResponseWriter, r *http.Request) {
	res := map[string]string{"message": "Membership Api", "version": "V.0.0.1"}
	json, err := json.Marshal(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	var reg Register

	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &reg)

	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		res := map[string]interface{}{"status": "Error", "message": "Error validation", "data": err.Error()}
		json, _ := json.Marshal(res)

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)

	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		res := map[string]interface{}{"status": "Error", "message": "Server error, unable to create your account.", "data": err.Error()}
		json, _ := json.Marshal(res)

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
		return
	}

	if reg.GroupID == 2 { // Borrower
		err = registerBorrower(reg.GroupID, reg.Email, hashedPassword, reg.Name, reg.CompanyName, reg.Address, reg.ProvinceID, reg.CityID, reg.PhoneNumber, reg.IdentityType, reg.IdentityFile, reg.NpwpFile, reg.SiupFile)

		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			res := map[string]interface{}{"status": "Error", "message": "Database error", "data": err.Error()}
			json, _ := json.Marshal(res)

			w.Header().Set("Content-Type", "application/json")
			w.Write(json)
			return
		}
	} else if reg.GroupID == 3 { // Investor
		err = registerInvestor(reg.GroupID, reg.Email, hashedPassword, reg.Name, reg.Gender, reg.BirthDate, reg.JobID, reg.Address, reg.ProvinceID, reg.CityID, reg.PhoneNumber, reg.IdentityType, reg.IdentityFile)

		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			res := map[string]interface{}{"status": "Error", "message": "Database error", "data": err.Error()}
			json, _ := json.Marshal(res)

			w.Header().Set("Content-Type", "application/json")
			w.Write(json)
			return
		}
	}

	res := map[string]interface{}{"status": "Ok", "message": "success", "data": nil}
	json, _ := json.Marshal(res)

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func registerBorrower(groupID int8, email string, password []byte, name string, companyName string, address string, provinceID int, cityID int, phoneNumber string, identityType int8, identityFile string, npwpFile string, siupFile string) error {
	var verificationCode string = generateVerificationCode()

	db := DBConnection()

	qUserAccount := `INSERT INTO user_accounts
		(group_id, email, password, verification_code) VALUES ($1, $2, $3, $4)
		RETURNING id`

	var userAccountID int

	err = db.QueryRow(qUserAccount, groupID, email, password, verificationCode).Scan(&userAccountID)

	qClient := `INSERT INTO clients
		(user_account_id, name, company_name, address, province_id, city_id, phone_number, identity_type, identity_file, npwp_file, siup_file)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = db.Query(qClient, userAccountID, name, companyName, address, provinceID, cityID, phoneNumber, identityType, identityFile, npwpFile, siupFile)
	defer db.Close()

	return err
}

func registerInvestor(groupID int8, email string, password []byte, name string, gender int8, birthDate string, jobID int8, address string, provinceID int, cityID int, phoneNumber string, identityType int8, identityFile string) error {
	var verificationCode string = generateVerificationCode()

	db := DBConnection()

	qUserAccount := `INSERT INTO user_accounts
		(group_id, email, password, verification_code) VALUES ($1, $2, $3, $4)
		RETURNING id`

	var userAccountID int

	err = db.QueryRow(qUserAccount, groupID, email, password, verificationCode).Scan(&userAccountID)

	qClient := `INSERT INTO clients
		(user_account_id, name, gender, birth_date, job_id, address, province_id, city_id, phone_number, identity_type, identity_file)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = db.Query(qClient, userAccountID, name, gender, birthDate, jobID, address, provinceID, cityID, phoneNumber, identityType, identityFile)
	defer db.Close()

	return err
}

func VerificationAccountPage(w http.ResponseWriter, r *http.Request) {
	var va VerificationAccount

	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &va)

	db := DBConnection()
	qAccount := `SELECT verification_code FROM user_accounts WHERE verification_code = $1 AND status = 0`
	err = db.QueryRow(qAccount, va.VerificationCode).Scan(&va.VerificationCode)

	if err != nil {
		defer db.Close()

		res := map[string]interface{}{"status": "Error", "message": err.Error(), "data": nil}
		json, _ := json.Marshal(res)

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
		return
	}

	qUpdateAccount := `UPDATE user_accounts SET status = 1 WHERE verification_code = $1`
	_, err = db.Query(qUpdateAccount, va.VerificationCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer db.Close()

	res := map[string]interface{}{"status": "Ok", "message": "success", "data": nil}
	json, _ := json.Marshal(res)

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	var login Login
	var email string
	var password []byte

	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &login)

	db := DBConnection()
	qAccount := `SELECT email, password FROM user_accounts WHERE email = $1 AND status = 1`
	err = db.QueryRow(qAccount, login.Email).Scan(&email, &password)

	if err != nil {
		defer db.Close()

		res := map[string]interface{}{"status": "Error", "message": err.Error(), "data": nil}
		json, _ := json.Marshal(res)

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(login.Password))

	if err != nil {
		defer db.Close()

		res := map[string]interface{}{"status": "Error", "message": err.Error(), "data": nil}
		json, _ := json.Marshal(res)

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
		return
	}

	defer db.Close()

	res := map[string]interface{}{"status": "Ok", "message": "success", "data": nil}
	json, _ := json.Marshal(res)

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func generateVerificationCode() string {
	var len int = 6

	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25))
	}
	return string(bytes)
}
