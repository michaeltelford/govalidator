// Package govalidator is package of validators and sanitizers for strings, structs and collections.
package govalidator

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	fieldsRequiredByDefault bool
	notNumberRegexp         = regexp.MustCompile("[^0-9]+")
	whiteSpacesAndMinus     = regexp.MustCompile("[\\s-]+")
	paramsRegexp            = regexp.MustCompile("\\(.*\\)$")
	errorsMap               map[string][]string
	tags                    tagMap
	msgs                    tagCustomMsgMap
)

const maxURLRuneCount = 2083
const minURLRuneCount = 3
const RF3339WithoutZone = "2006-01-02T15:04:05"

// SetFieldsRequiredByDefault causes validation to fail when struct fields
// do not include validations or are not explicitly marked as exempt (using `valid:"-"` or `valid:"email,optional"`).
// This struct definition will fail govalidator.ValidateStruct() (and the field values do not matter):
//     type exampleStruct struct {
//         Name  string ``
//         Email string `valid:"email"`
// This, however, will only fail when Email is empty or an invalid email address:
//     type exampleStruct2 struct {
//         Name  string `valid:"-"`
//         Email string `valid:"email"`
// Lastly, this will only fail when Email is an invalid email address but not when it's empty:
//     type exampleStruct2 struct {
//         Name  string `valid:"-"`
//         Email string `valid:"email,optional"`
func SetFieldsRequiredByDefault(value bool) {
	fieldsRequiredByDefault = value
}

// IsEmail check if the string is an email.
func IsEmail(str string) bool {
	// TODO uppercase letters are not supported
	return rxEmail.MatchString(str)
}

// IsExistingEmail checks if the string is an email of existing domain.
// Requires a network/Internet connection. Func and tests will fail otherwise.
func IsExistingEmail(email string) bool {
	if len(email) < 6 || len(email) > 254 {
		return false
	}

	at := strings.LastIndex(email, "@")
	if at <= 0 || at > len(email)-3 {
		return false
	}

	user := email[:at]
	host := email[at+1:]
	if len(user) > 64 {
		return false
	}

	if userDotRegexp.MatchString(user) || !userRegexp.MatchString(user) || !hostRegexp.MatchString(host) {
		return false
	}

	switch host {
	case "localhost", "example.com":
		return true
	}

	if _, err := net.LookupMX(host); err != nil {
		if _, err := net.LookupIP(host); err != nil {
			return false
		}
	}

	return true
}

// IsURL check if the string is an URL.
func IsURL(str string) bool {
	if str == "" || utf8.RuneCountInString(str) >= maxURLRuneCount || len(str) <= minURLRuneCount || strings.HasPrefix(str, ".") {
		return false
	}
	strTemp := str
	if strings.Index(str, ":") >= 0 && strings.Index(str, "://") == -1 {
		// support no indicated urlscheme but with colon for port number
		// http:// is appended so url.Parse will succeed, strTemp used so it does not impact rxURL.MatchString
		strTemp = "http://" + str
	}
	u, err := url.Parse(strTemp)
	if err != nil {
		return false
	}
	if strings.HasPrefix(u.Host, ".") {
		return false
	}
	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
		return false
	}
	return rxURL.MatchString(str)
}

// IsEmptyString returns true if empty.
func IsEmptyString(str string) bool {
	return strings.Trim(str, ` `) == ``
}

// IsNonEmptyString returns false if empty.
func IsNonEmptyString(str string) bool {
	return strings.Trim(str, ` `) != ``
}

// IsBoolean returns true if value can be converted to boolean.
func IsBoolean(str string) bool {
	if _, err := ToBoolean(str); err != nil {
		return false
	}
	return true
}

// IsRequestURL check if the string rawurl, assuming
// it was received in an HTTP request, is a valid
// URL confirm to RFC 3986
func IsRequestURL(rawurl string) bool {
	url, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return false //Couldn't even parse the rawurl
	}
	if len(url.Scheme) == 0 {
		return false //No Scheme found
	}
	return true
}

// IsRequestURI check if the string rawurl, assuming
// it was received in an HTTP request, is an
// absolute URI or an absolute path.
func IsRequestURI(rawurl string) bool {
	_, err := url.ParseRequestURI(rawurl)
	return err == nil
}

// IsAlpha check if the string contains only letters (a-zA-Z). Empty string is valid.
func IsAlpha(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxAlpha.MatchString(str)
}

//IsUTFLetter check if the string contains only unicode letter characters.
//Similar to IsAlpha but for all languages. Empty string is valid.
func IsUTFLetter(str string) bool {
	if IsNull(str) {
		return true
	}

	for _, c := range str {
		if !unicode.IsLetter(c) {
			return false
		}
	}
	return true

}

// IsAlphanumeric check if the string contains only letters and numbers. Empty string is valid.
func IsAlphanumeric(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxAlphanumeric.MatchString(str)
}

// IsUTFLetterNumeric check if the string contains only unicode letters and numbers. Empty string is valid.
func IsUTFLetterNumeric(str string) bool {
	if IsNull(str) {
		return true
	}
	for _, c := range str {
		if !unicode.IsLetter(c) && !unicode.IsNumber(c) { //letters && numbers are ok
			return false
		}
	}
	return true

}

// IsNumeric check if the string contains only numbers. Empty string is valid.
func IsNumeric(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxNumeric.MatchString(str)
}

// IsUTFNumeric check if the string contains only unicode numbers of any kind.
// Numbers can be 0-9 but also Fractions ¾,Roman Ⅸ and Hangzhou 〩. Empty string is valid.
func IsUTFNumeric(str string) bool {
	if IsNull(str) {
		return true
	}
	if strings.IndexAny(str, "+-") > 0 {
		return false
	}
	if len(str) > 1 {
		str = strings.TrimPrefix(str, "-")
		str = strings.TrimPrefix(str, "+")
	}
	for _, c := range str {
		if unicode.IsNumber(c) == false { //numbers && minus sign are ok
			return false
		}
	}
	return true

}

// IsUTFDigit check if the string contains only unicode radix-10 decimal digits. Empty string is valid.
func IsUTFDigit(str string) bool {
	if IsNull(str) {
		return true
	}
	if strings.IndexAny(str, "+-") > 0 {
		return false
	}
	if len(str) > 1 {
		str = strings.TrimPrefix(str, "-")
		str = strings.TrimPrefix(str, "+")
	}
	for _, c := range str {
		if !unicode.IsDigit(c) { //digits && minus sign are ok
			return false
		}
	}
	return true

}

// IsHexadecimal check if the string is a hexadecimal number.
func IsHexadecimal(str string) bool {
	return rxHexadecimal.MatchString(str)
}

// IsHexcolor check if the string is a hexadecimal color.
func IsHexcolor(str string) bool {
	return rxHexcolor.MatchString(str)
}

// IsRGBcolor check if the string is a valid RGB color in form rgb(RRR, GGG, BBB).
func IsRGBcolor(str string) bool {
	return rxRGBcolor.MatchString(str)
}

// IsLowerCase check if the string is lowercase. Empty string is valid.
func IsLowerCase(str string) bool {
	if IsNull(str) {
		return true
	}
	return str == strings.ToLower(str)
}

// IsUpperCase check if the string is uppercase. Empty string is valid.
func IsUpperCase(str string) bool {
	if IsNull(str) {
		return true
	}
	return str == strings.ToUpper(str)
}

// HasLowerCase check if the string contains at least 1 lowercase. Empty string is valid.
func HasLowerCase(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxHasLowerCase.MatchString(str)
}

// HasUpperCase check if the string contians as least 1 uppercase. Empty string is valid.
func HasUpperCase(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxHasUpperCase.MatchString(str)
}

// IsInt check if the string is an integer. Empty string is valid.
func IsInt(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxInt.MatchString(str)
}

// IsFloat check if the string is a float.
func IsFloat(str string) bool {
	return str != "" && rxFloat.MatchString(str)
}

// IsDivisibleBy check if the string is a number that's divisible by another.
// If second argument is not valid integer or zero, it's return false.
// Otherwise, if first argument is not valid integer or zero, it's return true (Invalid string converts to zero).
func IsDivisibleBy(str, num string) bool {
	f, _ := ToFloat(str)
	p := int64(f)
	q, _ := ToInt(num)
	if q == 0 {
		return false
	}
	return (p == 0) || (p%q == 0)
}

// IsNull check if the string is null.
func IsNull(str string) bool {
	return len(str) == 0
}

// IsByteLength check if the string's length (in bytes) falls in a range.
func IsByteLength(str string, min, max int) bool {
	return len(str) >= min && len(str) <= max
}

// IsUUIDv3 check if the string is a UUID version 3.
func IsUUIDv3(str string) bool {
	return rxUUID3.MatchString(str)
}

// IsUUIDv4 check if the string is a UUID version 4.
func IsUUIDv4(str string) bool {
	return rxUUID4.MatchString(str)
}

// IsUUIDv5 check if the string is a UUID version 5.
func IsUUIDv5(str string) bool {
	return rxUUID5.MatchString(str)
}

// IsUUID check if the string is a UUID (version 3, 4 or 5).
func IsUUID(str string) bool {
	return rxUUID.MatchString(str)
}

// IsCreditCard check if the string is a credit card.
func IsCreditCard(str string) bool {
	sanitized := notNumberRegexp.ReplaceAllString(str, "")
	if !rxCreditCard.MatchString(sanitized) {
		return false
	}
	var sum int64
	var digit string
	var tmpNum int64
	var shouldDouble bool
	for i := len(sanitized) - 1; i >= 0; i-- {
		digit = sanitized[i:(i + 1)]
		tmpNum, _ = ToInt(digit)
		if shouldDouble {
			tmpNum *= 2
			if tmpNum >= 10 {
				sum += ((tmpNum % 10) + 1)
			} else {
				sum += tmpNum
			}
		} else {
			sum += tmpNum
		}
		shouldDouble = !shouldDouble
	}

	if sum%10 == 0 {
		return true
	}
	return false
}

// IsISBN10 check if the string is an ISBN version 10.
func IsISBN10(str string) bool {
	return IsISBN(str, 10)
}

// IsISBN13 check if the string is an ISBN version 13.
func IsISBN13(str string) bool {
	return IsISBN(str, 13)
}

// IsISBN check if the string is an ISBN (version 10 or 13).
// If version value is not equal to 10 or 13, it will be check both variants.
func IsISBN(str string, version int) bool {
	sanitized := whiteSpacesAndMinus.ReplaceAllString(str, "")
	var checksum int32
	var i int32
	if version == 10 {
		if !rxISBN10.MatchString(sanitized) {
			return false
		}
		for i = 0; i < 9; i++ {
			checksum += (i + 1) * int32(sanitized[i]-'0')
		}
		if sanitized[9] == 'X' {
			checksum += 10 * 10
		} else {
			checksum += 10 * int32(sanitized[9]-'0')
		}
		if checksum%11 == 0 {
			return true
		}
		return false
	} else if version == 13 {
		if !rxISBN13.MatchString(sanitized) {
			return false
		}
		factor := []int32{1, 3}
		for i = 0; i < 12; i++ {
			checksum += factor[i%2] * int32(sanitized[i]-'0')
		}
		if (int32(sanitized[12]-'0'))-((10-(checksum%10))%10) == 0 {
			return true
		}
		return false
	}
	return IsISBN(str, 10) || IsISBN(str, 13)
}

// IsJSON check if the string is valid JSON (note: uses json.Unmarshal).
func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// IsMultibyte check if the string contains one or more multibyte chars. Empty string is valid.
func IsMultibyte(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxMultibyte.MatchString(str)
}

// IsASCII check if the string contains ASCII chars only. Empty string is valid.
func IsASCII(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxASCII.MatchString(str)
}

// IsPrintableASCII check if the string contains printable ASCII chars only. Empty string is valid.
func IsPrintableASCII(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxPrintableASCII.MatchString(str)
}

// IsFullWidth check if the string contains any full-width chars. Empty string is valid.
func IsFullWidth(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxFullWidth.MatchString(str)
}

// IsHalfWidth check if the string contains any half-width chars. Empty string is valid.
func IsHalfWidth(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxHalfWidth.MatchString(str)
}

// IsVariableWidth check if the string contains a mixture of full and half-width chars. Empty string is valid.
func IsVariableWidth(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxHalfWidth.MatchString(str) && rxFullWidth.MatchString(str)
}

// IsBase64 check if a string is base64 encoded.
func IsBase64(str string) bool {
	return rxBase64.MatchString(str)
}

// IsFilePath check is a string is Win or Unix file path and returns it's type.
func IsFilePath(str string) (bool, int) {
	if rxWinPath.MatchString(str) {
		//check windows path limit see:
		//  http://msdn.microsoft.com/en-us/library/aa365247(VS.85).aspx#maxpath
		if len(str[3:]) > 32767 {
			return false, Win
		}
		return true, Win
	} else if rxUnixPath.MatchString(str) {
		return true, Unix
	}
	return false, Unknown
}

// IsDataURI checks if a string is base64 encoded data URI such as an image
func IsDataURI(str string) bool {
	dataURI := strings.Split(str, ",")
	if !rxDataURI.MatchString(dataURI[0]) {
		return false
	}
	return IsBase64(dataURI[1])
}

// IsISO3166Alpha2 checks if a string is valid two-letter country code
func IsISO3166Alpha2(str string) bool {
	for _, entry := range ISO3166List {
		if str == entry.Alpha2Code {
			return true
		}
	}
	return false
}

// IsISO3166Alpha3 checks if a string is valid three-letter country code
func IsISO3166Alpha3(str string) bool {
	for _, entry := range ISO3166List {
		if str == entry.Alpha3Code {
			return true
		}
	}
	return false
}

// IsISO693Alpha2 checks if a string is valid two-letter language code
func IsISO693Alpha2(str string) bool {
	for _, entry := range ISO693List {
		if str == entry.Alpha2Code {
			return true
		}
	}
	return false
}

// IsISO693Alpha3b checks if a string is valid three-letter language code
func IsISO693Alpha3b(str string) bool {
	for _, entry := range ISO693List {
		if str == entry.Alpha3bCode {
			return true
		}
	}
	return false
}

// IsDNSName will validate the given string as a DNS name
func IsDNSName(str string) bool {
	if str == "" || len(strings.Replace(str, ".", "", -1)) > 255 {
		// constraints already violated
		return false
	}
	return !IsIP(str) && rxDNSName.MatchString(str)
}

// IsHash checks if a string is a hash of type algorithm.
// Algorithm is one of ['md4', 'md5', 'sha1', 'sha256', 'sha384', 'sha512', 'ripemd128', 'ripemd160', 'tiger128', 'tiger160', 'tiger192', 'crc32', 'crc32b']
func IsHash(str string, algorithm string) bool {
	len := "0"
	algo := strings.ToLower(algorithm)

	if algo == "crc32" || algo == "crc32b" {
		len = "8"
	} else if algo == "md5" || algo == "md4" || algo == "ripemd128" || algo == "tiger128" {
		len = "32"
	} else if algo == "sha1" || algo == "ripemd160" || algo == "tiger160" {
		len = "40"
	} else if algo == "tiger192" {
		len = "48"
	} else if algo == "sha256" {
		len = "64"
	} else if algo == "sha384" {
		len = "96"
	} else if algo == "sha512" {
		len = "128"
	} else {
		return false
	}

	return Matches(str, "^[a-f0-9]{"+len+"}$")
}

// IsDialString validates the given string for usage with the various Dial() functions
func IsDialString(str string) bool {

	if h, p, err := net.SplitHostPort(str); err == nil && h != "" && p != "" && (IsDNSName(h) || IsIP(h)) && IsPort(p) {
		return true
	}

	return false
}

// IsIP checks if a string is either IP version 4 or 6.
func IsIP(str string) bool {
	return net.ParseIP(str) != nil
}

// IsPort checks if a string represents a valid port
func IsPort(str string) bool {
	if i, err := strconv.Atoi(str); err == nil && i > 0 && i < 65536 {
		return true
	}
	return false
}

// IsIPv4 check if the string is an IP version 4.
func IsIPv4(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ".")
}

// IsIPv6 check if the string is an IP version 6.
func IsIPv6(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ":")
}

// IsCIDR check if the string is an valid CIDR notiation (IPV4 & IPV6)
func IsCIDR(str string) bool {
	_, _, err := net.ParseCIDR(str)
	return err == nil
}

// IsMAC check if a string is valid MAC address.
// Possible MAC formats:
// 01:23:45:67:89:ab
// 01:23:45:67:89:ab:cd:ef
// 01-23-45-67-89-ab
// 01-23-45-67-89-ab-cd-ef
// 0123.4567.89ab
// 0123.4567.89ab.cdef
func IsMAC(str string) bool {
	_, err := net.ParseMAC(str)
	return err == nil
}

// IsHost checks if the string is a valid IP (both v4 and v6) or a valid DNS name
func IsHost(str string) bool {
	return IsIP(str) || IsDNSName(str)
}

// IsMongoID check if the string is a valid hex-encoded representation of a MongoDB ObjectId.
func IsMongoID(str string) bool {
	return rxHexadecimal.MatchString(str) && (len(str) == 24)
}

// IsLatitude check if a string is valid latitude.
func IsLatitude(str string) bool {
	return rxLatitude.MatchString(str)
}

// IsLongitude check if a string is valid longitude.
func IsLongitude(str string) bool {
	return rxLongitude.MatchString(str)
}

// IsRsaPublicKey check if a string is valid public key with provided length
func IsRsaPublicKey(str string, keylen int) bool {
	bb := bytes.NewBufferString(str)
	pemBytes, err := ioutil.ReadAll(bb)
	if err != nil {
		return false
	}
	block, _ := pem.Decode(pemBytes)
	if block != nil && block.Type != "PUBLIC KEY" {
		return false
	}
	var der []byte

	if block != nil {
		der = block.Bytes
	} else {
		der, err = base64.StdEncoding.DecodeString(str)
		if err != nil {
			return false
		}
	}

	key, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return false
	}
	pubkey, ok := key.(*rsa.PublicKey)
	if !ok {
		return false
	}
	bitlen := len(pubkey.N.Bytes()) * 8
	return bitlen == int(keylen)
}

func toJSONName(tag string) string {
	if tag == "" {
		return ""
	}

	// JSON name always comes first. If there's no options then split[0] is
	// JSON name, if JSON name is not set, then split[0] is an empty string.
	split := strings.SplitN(tag, ",", 2)

	name := split[0]

	// However it is possible that the field is skipped when
	// (de-)serializing from/to JSON, in which case assume that there is no
	// tag name to use
	if name == "-" {
		return ""
	}
	return name
}

// Validate a struct using its `valid` field tags.
// Returns an isValid boolean and all validation errors found listed in a map
// for easy post processing e.g. JSON marshalling etc.
func Validate(i interface{}) (bool, map[string]map[string][]string) {
	errorsMap = make(map[string][]string, 0)
	valid, _ := validateStruct(i)
	removeDuplicateErrors()
	return valid, allErrors()
}

// validateStruct uses `valid` field tags as validation rules.
// Returns an isValid boolean and the first validation error found.
func validateStruct(s interface{}) (bool, error) {
	if s == nil {
		return true, nil
	}

	result := true
	var err error

	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// We only accept structs.
	if val.Kind() != reflect.Struct {
		return false, fmt.Errorf("function only accepts structs; got %s", val.Kind())
	}

	var errs Errors
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		validTag := typeField.Tag.Get(tagName)

		if typeField.PkgPath != "" {
			continue // Private field.
		}

		structResult := true

		// If `valid` isn't "-" and concrete field is a struct.
		if validTag != "-" && (valueField.Kind() == reflect.Struct ||
			(valueField.Kind() == reflect.Ptr && valueField.Elem().Kind() == reflect.Struct)) {
			var err error
			structResult, err = validateStruct(valueField.Interface())
			if err != nil {
				errs = append(errs, NewError(err))
			}
		}

		parseTagIntoMap(validTag)
		resultField, err2 := validateField(valueField, typeField, val, true)
		if err2 != nil {
			// Replace field name with JSON name if present.
			jsonTag := toJSONName(typeField.Tag.Get("json"))
			if jsonTag != "" {
				switch jsonError := err2.(type) {
				case Error:
					jsonError.Name = jsonTag
					err2 = jsonError
				case Errors:
					for i2, err3 := range jsonError {
						switch customErr := interface{}(err3).(type) {
						case Error:
							customErr.Name = jsonTag
							jsonError[i2] = customErr
						}
					}
					err2 = jsonError
				}
			}

			errs = append(errs, NewError(err2))
		}

		result = result && resultField && structResult
	}

	if len(errs) > 0 {
		err = errs
	}
	return result, err
}

// validateField runs all validators for a single struct field.
// v is struct field value, t is struct field type and o is the full struct (value).
func validateField(v reflect.Value, t reflect.StructField, o reflect.Value, isRootType bool) (isValid bool, resultErr error) {
	var validResult bool
	var err error
	var firstErr error

	if !v.IsValid() {
		return false, nil
	}

	tag := t.Tag.Get(tagName) // `valid`
	jsonTag := t.Tag.Get(`json`)

	// Check if the field should be ignored: `valid:""` or `valid:"-"` tags.
	switch tag {
	case "":
		if !fieldsRequiredByDefault {
			return true, nil
		}
		e := Error{t.Name, fmt.Errorf("All fields are required to at least have one validation defined"), false, "required"}
		appendErrorsMap(jsonTag, e)
		return false, e
	case "-":
		return true, nil
	}

	// Presence validation; if the value is empty, process the `required`
	// and `optional` tags otherwise process the `forbidden` tag.
	if isEmptyValue(v) {
		// Process `required` and `optional` tags.
		if tempIsValid, tempError := checkRequired(v, t, msgs); !tempIsValid && tempError != nil {
			validResult = false
			err = tempError
			if firstErr == nil {
				firstErr = err
			}
		} else if _, isOptional := msgs["optional"]; tempIsValid && tempError == nil && isOptional {
			// At this point, we know the value is empty and the optional tag
			// is present so don't bother with other validators (which are
			// only run if non zero value). Return valid=true.
			return true, nil
		}
	} else {
		// Process `forbidden` tag.
		if tempIsValid, tempError := checkForbidden(v, t, msgs); !tempIsValid && tempError != nil {
			validResult = false
			err = tempError
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	var customTypeErrors Errors
	for _, tag := range tags {
		customErrorMessage := msgs[tag]
		if validatefunc, ok := CustomTypeTagMap.Get(tag); ok {
			deleteTagAndMsg(tag)

			if result := validatefunc(v.Interface(), o.Interface()); !result {
				if len(customErrorMessage) > 0 {
					customTypeErrors = append(customTypeErrors, Error{Name: t.Name, Err: fmt.Errorf(customErrorMessage), CustomErrorMessageExists: true, Validator: stripParams(tag)})
					continue
				}
				customTypeErrors = append(customTypeErrors, Error{Name: t.Name, Err: fmt.Errorf("%s does not validate as %s", fmt.Sprint(v), tag), CustomErrorMessageExists: false, Validator: stripParams(tag)})
			}
		}
	}
	if len(customTypeErrors) > 0 {
		for _, customErr := range customTypeErrors {
			// Add to the map of all validation errors in the struct.
			if firstErr == nil {
				firstErr = customErr.Err
			}
			appendErrorsMap(jsonTag, customErr)
		}
		return false, customTypeErrors
	}

	if isRootType {
		// Ensure that we've checked the value by all specified validators before report that the value is valid.
		defer func() {
			deleteTagAndMsg("optional")
			deleteTagAndMsg("required")
			deleteTagAndMsg("forbidden")

			if isValid && resultErr == nil && len(tags) != 0 {
				for _, validator := range tags {
					isValid = false
					resultErr = Error{t.Name, fmt.Errorf(
						"The following validator is invalid or can't be applied to the field: %q", validator), false, stripParams(validator)}
					return
				}
			}
		}()
	}

	switch v.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.String:

		// for each tag option check the map of validator functions
		for _, tag := range tags {
			validatorSpec := tag
			customErrorMessage := msgs[tag]

			var negate bool
			validator := validatorSpec
			customMsgExists := len(customErrorMessage) > 0

			// Check whether the tag looks like '!something' or 'something'
			if validator[0] == '!' {
				validator = validator[1:]
				negate = true
			}

			// Check for param validators
			for key, value := range ParamTagRegexMap {
				ps := value.FindStringSubmatch(validator)
				if len(ps) == 0 {
					continue
				}

				validatefunc, ok := ParamTagMap[key]
				if !ok {
					continue
				}

				deleteTagAndMsg(tag)

				switch v.Kind() {
				case reflect.String,
					reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
					reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
					reflect.Float32, reflect.Float64:

					field := fmt.Sprint(v) // make value into string, then validate with regex
					if result := validatefunc(field, ps[1:]...); (!result && !negate) || (result && negate) {
						if customMsgExists {
							validResult, err = false, Error{t.Name, fmt.Errorf(customErrorMessage), customMsgExists, stripParams(validatorSpec)}
						} else {
							validResult, err = false, Error{t.Name, fmt.Errorf("%s does not validate as %s", field, validator), customMsgExists, stripParams(validatorSpec)}
						}
						if negate {
							validResult, err = false, Error{t.Name, fmt.Errorf("%s does validate as %s", field, validator), customMsgExists, stripParams(validatorSpec)}
						}
					}
				default:
					// type not yet supported, fail
					validResult, err = false, Error{t.Name, fmt.Errorf("Validator %s doesn't support kind %s", validator, v.Kind()), false, stripParams(validatorSpec)}
				}
			}

			if validatefunc, ok := TagMap[validator]; ok {
				deleteTagAndMsg(tag)

				switch v.Kind() {
				case reflect.String:
					field := fmt.Sprint(v) // make value into string, then validate with regex
					if result := validatefunc(field); !result && !negate || result && negate {
						if customMsgExists {
							validResult, err = false, Error{t.Name, fmt.Errorf(customErrorMessage), customMsgExists, stripParams(validatorSpec)}
						} else {
							validResult, err = false, Error{t.Name, fmt.Errorf("%s does not validate as %s", field, validator), customMsgExists, stripParams(validatorSpec)}
						}
						if negate {
							validResult, err = false, Error{t.Name, fmt.Errorf("%s does validate as %s", field, validator), customMsgExists, stripParams(validatorSpec)}
						}
					}
				default:
					//Not Yet Supported Types (Fail here!)
					err := fmt.Errorf("Validator %s doesn't support kind %s for value %v", validator, v.Kind(), v)
					validResult, err = false, Error{t.Name, err, false, stripParams(validatorSpec)}
				}
			}

			// Add to the map of all validation errors in the struct.
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				appendErrorsMap(jsonTag, NewError(err))
			}
		}

		if !validResult && firstErr != nil {
			return false, firstErr
		}

		return true, nil
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return false, &UnsupportedTypeError{v.Type()}
		}
		var sv stringValues
		sv = v.MapKeys()
		sort.Sort(sv)
		result := true
		for _, k := range sv {
			var resultItem bool
			var err error
			if v.MapIndex(k).Kind() != reflect.Struct {
				resultItem, err = validateField(v.MapIndex(k), t, o, false)
				if err != nil {
					return false, err
				}
			} else {
				resultItem, err = validateStruct(v.MapIndex(k).Interface())
				if err != nil {
					return false, err
				}
			}
			result = result && resultItem
		}
		return result, nil
	case reflect.Slice, reflect.Array:
		result := true
		for i := 0; i < v.Len(); i++ {
			var resultItem bool
			var err error
			if v.Index(i).Kind() != reflect.Struct {
				resultItem, err = validateField(v.Index(i), t, o, false)
				if err != nil {
					return false, err
				}
			} else {
				resultItem, err = validateStruct(v.Index(i).Interface())
				if err != nil {
					return false, err
				}
			}
			result = result && resultItem
		}
		return result, nil
	case reflect.Interface:
		// If the value is an interface then encode its element
		if v.IsNil() {
			return true, nil
		}
		return validateStruct(v.Interface())
	case reflect.Ptr:
		// If the value is a pointer then check its element
		if v.IsNil() {
			return true, nil
		}
		return validateField(v.Elem(), t, o, false)
	case reflect.Struct:
		return validateStruct(v.Interface())
	default:
		return false, &UnsupportedTypeError{v.Type()}
	}
}

func allErrors() map[string]map[string][]string {
	return map[string]map[string][]string{"errors": errorsMap}
}

func appendErrorsMap(attr string, err Error) {
	if errorsMap == nil {
		return
	}

	attr = toJSONName(attr)
	errMsg := err.Error()

	errorsMap[attr] = append(errorsMap[attr], errMsg)
}

func removeDuplicateErrors() {
	for attr, errs := range errorsMap {
		errorsMap[attr] = removeDuplicates(errs)
	}
}

func removeDuplicates(xs []string) []string {
	found := make(map[string]bool)
	j := 0
	for i, x := range xs {
		if !found[x] {
			found[x] = true
			xs[j] = xs[i]
			j++
		}
	}
	return xs[:j]
}

// parseTagIntoMap parses all valid:`` tags for a single struct field into a []
// string (maintaining order) and then parses a struct tag of
// `valid:required~Some error message,length(2|3)` into message
// map[string]string{"required": "Some error message", "length(2|3)": ""}
func parseTagIntoMap(tag string) {
	tags = make(tagMap, 0)
	msgs = make(tagCustomMsgMap, 0)

	options := strings.Split(tag, ",")

	for _, option := range options {
		option = strings.TrimSpace(option)

		validationOptions := strings.Split(option, "~")
		if !isValidTag(validationOptions[0]) {
			continue
		}

		tags = append(tags, validationOptions[0])

		if len(validationOptions) == 2 {
			msgs[validationOptions[0]] = validationOptions[1]
		} else {
			msgs[validationOptions[0]] = ""
		}
	}
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("\\'\"!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		default:
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}

// IsSSN will validate the given string as a U.S. Social Security Number
func IsSSN(str string) bool {
	if str == "" || len(str) != 11 {
		return false
	}
	return rxSSN.MatchString(str)
}

// IsSemver check if string is valid semantic version
func IsSemver(str string) bool {
	return rxSemver.MatchString(str)
}

// IsTime check if string is valid according to given format
func IsTime(str string, format string) bool {
	_, err := time.Parse(format, str)
	return err == nil
}

// IsRFC3339 check if string is valid timestamp value according to RFC3339
func IsRFC3339(str string) bool {
	return IsTime(str, time.RFC3339)
}

// IsRFC3339WithoutZone check if string is valid timestamp value according to RFC3339 which excludes the timezone.
func IsRFC3339WithoutZone(str string) bool {
	return IsTime(str, RF3339WithoutZone)
}

// IsISO4217 check if string is valid ISO currency code
func IsISO4217(str string) bool {
	for _, currency := range ISO4217List {
		if str == currency {
			return true
		}
	}

	return false
}

// ByteLength check string's length
func ByteLength(str string, params ...string) bool {
	if len(params) == 2 {
		min, _ := ToInt(params[0])
		max, _ := ToInt(params[1])
		return len(str) >= int(min) && len(str) <= int(max)
	}

	return false
}

// RuneLength check string's length
// Alias for StringLength
func RuneLength(str string, params ...string) bool {
	return StringLength(str, params...)
}

// IsRsaPub check whether string is valid RSA key
// Alias for IsRsaPublicKey
func IsRsaPub(str string, params ...string) bool {
	if len(params) == 1 {
		len, _ := ToInt(params[0])
		return IsRsaPublicKey(str, int(len))
	}

	return false
}

// StringMatches checks if a string matches a given pattern.
func StringMatches(s string, params ...string) bool {
	if len(params) == 1 {
		pattern := params[0]
		return Matches(s, pattern)
	}
	return false
}

// StringLength check string's length (including multi byte strings)
func StringLength(str string, params ...string) bool {

	if len(params) == 2 {
		strLength := utf8.RuneCountInString(str)
		min, _ := ToInt(params[0])
		max, _ := ToInt(params[1])
		return strLength >= int(min) && strLength <= int(max)
	}

	return false
}

// Range check string's length
func Range(str string, params ...string) bool {
	if len(params) == 2 {
		value, _ := ToFloat(str)
		min, _ := ToFloat(params[0])
		max, _ := ToFloat(params[1])
		return InRange(value, min, max)
	}

	return false
}

func isInRaw(str string, params ...string) bool {
	if len(params) == 1 {
		rawParams := params[0]

		parsedParams := strings.Split(rawParams, "|")

		return IsIn(str, parsedParams...)
	}

	return false
}

// IsIn check if string str is a member of the set of strings params
func IsIn(str string, params ...string) bool {
	for _, param := range params {
		if str == param {
			return true
		}
	}

	return false
}

// Process `required` and `optional` tags if present.
func checkRequired(v reflect.Value, t reflect.StructField, options tagCustomMsgMap) (bool, error) {
	if requiredOption, isRequired := options["required"]; isRequired {
		if len(requiredOption) > 0 {
			return false, Error{t.Name, fmt.Errorf(requiredOption), true, "required"}
		}
		return false, Error{t.Name, fmt.Errorf("non zero value required"), false, "required"}
	} else if _, isOptional := options["optional"]; fieldsRequiredByDefault && !isOptional {
		return false, Error{t.Name, fmt.Errorf("Missing required field"), false, "required"}
	}
	// not required and empty is valid
	return true, nil
}

// Process `forbidden` tag if present.
func checkForbidden(v reflect.Value, t reflect.StructField, options tagCustomMsgMap) (bool, error) {
	if option, found := options[`forbidden`]; found {
		if len(option) > 0 {
			return false, Error{t.Name, fmt.Errorf(option), true, `forbidden`}
		}
		return false, Error{t.Name, fmt.Errorf(`Illegal attribute`), false, `forbidden`}
	}
	return true, nil
}

func deleteTagAndMsg(tag string) {
	delete(msgs, tag)
	for i, t := range tags {
		if t == tag {
			tags = append(tags[:i], tags[i+1:]...)
			return
		}
	}
}

func stripParams(validatorString string) string {
	return paramsRegexp.ReplaceAllString(validatorString, "")
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String, reflect.Array:
		return v.Len() == 0
	case reflect.Map, reflect.Slice:
		return v.Len() == 0 || v.IsNil()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

// ErrorByField returns error for specified field of the struct
// validated by ValidateStruct or empty string if there are no errors
// or this field doesn't exists or doesn't have any errors.
func ErrorByField(e error, field string) string {
	if e == nil {
		return ""
	}
	return ErrorsByField(e)[field]
}

// ErrorsByField returns map of errors of the struct validated
// by ValidateStruct or empty map if there are no errors.
func ErrorsByField(e error) map[string]string {
	m := make(map[string]string)
	if e == nil {
		return m
	}
	// prototype for ValidateStruct

	switch e.(type) {
	case Error:
		m[e.(Error).Name] = e.(Error).Err.Error()
	case Errors:
		for _, item := range e.(Errors).Errors() {
			n := ErrorsByField(item)
			for k, v := range n {
				m[k] = v
			}
		}
	}

	return m
}

// Error returns string equivalent for reflect.Type
func (e *UnsupportedTypeError) Error() string {
	return "validator: unsupported type: " + e.Type.String()
}

func (sv stringValues) Len() int           { return len(sv) }
func (sv stringValues) Swap(i, j int)      { sv[i], sv[j] = sv[j], sv[i] }
func (sv stringValues) Less(i, j int) bool { return sv.get(i) < sv.get(j) }
func (sv stringValues) get(i int) string   { return sv[i].String() }
