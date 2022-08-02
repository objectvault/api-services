// cSpell:ignore bson, gonic, paulo ferreira, userro
package common

/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"log"
	"net/http"
)

func codesOk(code int) (int, string) {
	switch code {
	case 1000:
		return http.StatusOK, "OK"
	case 1001:
		return http.StatusOK, "Please Login!"
	case 1002:
		return http.StatusOK, "Logged Out!"
	case 1003:
		return http.StatusOK, "Logged In!"
	case 1099:
		return http.StatusOK, "Contact System Administrator."
	case 1998: // TODO Set Proper Error Code
		return http.StatusOK, "TO BE Implemented - Success Code"
	default:
		log.Printf("[codesOk] Unrecognized Status Code [%d]\n", code)
		return http.StatusOK, "Unknown Reason"
	}
}

func codesWarnings(code int) (int, string) {
	switch code {
	// 2000 - 2099 : User Related Warnings
	case 2001: // Update Profile with No Changes
		return http.StatusOK, "There was nothing to change!"
		// 2100 - 2199 : Organization Related Warnings
		// 2200 - 2299 : Store Related Warnings
		// 2300 - 2399 : Store Related Warnings
		// 2400 - 2499 : Invitation Related Messages
	case 2490: // Failed to Queue Message
		return http.StatusOK, "Failed to Send Invitation. Retry!"
	// 2900 - 2999 : System Related Warning
	case 2998: // TODO Set Proper Error Code
		return http.StatusOK, "TO BE Implemented - Warning Code"
	default:
		log.Printf("[codesWarnings] Unrecognized Status Code [%d]\n", code)
		return http.StatusOK, "Unknown Reason"
	}
}

func codesProcessingErrors(code int) (int, string) {
	switch code {
	// 3000 - 3099 Session Creation or Validation
	case 3000:
		return http.StatusBadRequest, "Not Logged In!"
	case 3001: // Invalid Login Credentials
		return http.StatusBadRequest, "Invalid Login Credentials"
	case 3002: // Invalid Session User
		return http.StatusBadRequest, "Session User Invalid"
	case 3003: // User Logged In
		return http.StatusBadRequest, "User Session Active"
	case 3004: // Session Not Registered
		return http.StatusBadRequest, "Not a Registered User Session"
	case 3010: // User is not Associated with a Company
		return http.StatusBadRequest, "Session User is not a Company User"
	case 3011: // User is Not Company Admin
		return http.StatusBadRequest, "Session User is not Company Administrator Account"
	// 3100 - 3199 API Parameter Validation
	case 3100:
		return http.StatusBadRequest, "Missing or Invalid API Parameters"
	// 3200 - 3299 Form Paramater Validation
	case 3200:
		return http.StatusBadRequest, "Missing or Invalid Form Parameters"
	case 3201:
		return http.StatusBadRequest, "No Valid Form Parameters Passed"
	// 3300 - 3399 URL Paramater Validation
	case 3300:
		return http.StatusBadRequest, "Missing or Invalid URL Parameters"
	case 3301:
		return http.StatusBadRequest, "No Valid URL Parameters Passed"
		// 3400 - 3499 License Errors
		// 3500 - 3599 License Instance Errors
	case 3500:
		return http.StatusBadRequest, "Missing or Invalid License Parameters"
	case 3998: // TODO Set Proper Error Code
		return http.StatusBadRequest, "TO BE Implemented - Processing Code"
	default:
		log.Printf("[codesProcessingErrors] Unrecognized Status Code [%d]\n", code)
		return http.StatusBadRequest, "Unknown Reason"
	}
}

func codesObjectErrors(code int) (int, string) {
	switch code {
	// 4000 - 4099 : User Related Errors
	case 4000: // User Does not Exist
		return http.StatusBadRequest, "User does not exist"
	case 4001: // User Inactive
		return http.StatusBadRequest, "User account inactive"
	case 4002: // User Locked Out
		return http.StatusBadRequest, "User account disabled"
	case 4003: // User Missing Required Roles
		return http.StatusBadRequest, "User Access Denied"
	case 4004: // Session User is same as Request User
		return http.StatusBadRequest, "Action Not Permitted on SELF"
	case 4010: // Alias Exists
		return http.StatusBadRequest, "Alias already Exists"
	case 4011: // Email Registered
		return http.StatusBadRequest, "Email already registered"
	case 4012: // User <--> Object Registration Exists
		return http.StatusBadRequest, "User already regsitered with object"
	// 4100 - 4199 : Organization Related Errors
	case 4100: // Organization Does not Exist
		return http.StatusBadRequest, "Organization does not exist"
	case 4101: // Action not allowed
		return http.StatusBadRequest, "Access Denied"
	// 4200 - 4299 : User Activation Related Errors
	case 4200: // Failed to Create User Activation Code
		return http.StatusBadRequest, "Activation does not exist"
	case 4201: // Activation Code is Invalid (serves for Expired Codes Too)
		return http.StatusBadRequest, "Invalid Activation Code"
	case 4202: // User Already Active
		return http.StatusBadRequest, "User already activated!"
	case 4203: // Failed to Create User Activation Code
		return http.StatusBadRequest, "Failed to Create/Send Activation Code. Retry!"
	// 4300 - 4399 : Invitation Related Error
	case 4300:
		return http.StatusBadRequest, "Invitation Accept, requires Session by the invitee!"
	case 4301:
		return http.StatusBadRequest, "Invitation Accept, requires Session and Password by the invitee!"
	case 4390:
		return http.StatusBadRequest, "Invalid Invitation ID!"
	case 4391:
		return http.StatusBadRequest, "Invitation Expired!"
	// 4400 - 4499 : Invitation Related Error
	case 4400:
		return http.StatusBadRequest, "Template does not exist"
		// 4500 - 4599 : Request Related Error
	case 4500:
		return http.StatusBadRequest, "Request does not exist"
	case 4591:
		return http.StatusBadRequest, "Request Expired!"
	case 4998: // TODO Set Proper Error Code
		return http.StatusInternalServerError, "TO BE Implemented - Object Code"
	default:
		log.Printf("[codesDatabaseErrors] Unrecognized Status Code [%d]\n", code)
		return http.StatusBadRequest, "Unknown Reason"
	}
}

func codesServerErrors(code int) (int, string) {
	switch code {
	// 5000 - 5099 : Server Session Errors
	case 5000:
		return http.StatusInternalServerError, "Failed to Create Session"
	case 5001:
		return http.StatusInternalServerError, "Failed to Clear Session"
	case 5010:
		return http.StatusInternalServerError, "Failed to Open Store"
	// 5100 - 5199 : Database Related Errors
	case 5100:
		return http.StatusInternalServerError, "Database Error"
	// 5200 - 5299 : Request Related Errors
	case 5200: // Generic Request Error
		return http.StatusBadRequest, "Invalid Request"
	case 5201: // Not a JSON Request Error
		return http.StatusBadRequest, "NOT a Valid JSON Request"
	case 5202: // Received JSON Request but is not Valid
		return http.StatusBadRequest, "JSON Request is Not Valid"
	// 5300 - 5399 : Queue Related Errors
	case 5300:
		return http.StatusInternalServerError, "General Message Queue Error"
	case 5301:
		return http.StatusInternalServerError, "Failed Sending Message"
	case 5302:
		return http.StatusInternalServerError, "Failed Connecting to Queue Server"
	case 5303:
		return http.StatusInternalServerError, "General Request Error"
	// 5900 - 5999 : Unexpected Server Errors
	case 5900: // Session Error
		return http.StatusInternalServerError, "Unexpected Server Error"
	case 5901: // JSON Conversion Error
		return http.StatusInternalServerError, "Error Converting to JSON"
	case 5920: // Queue Message Creation Error
		return http.StatusInternalServerError, "System Error Creating Queue Message"
	case 5921: // Queue Message Publish Error
		return http.StatusInternalServerError, "System Error Publishing Queue Message"
	case 5998: // TODO Set Proper Error Code
		return http.StatusInternalServerError, "TO BE Implemented - Error Code"
	case 5999: // TODO Error
		return http.StatusInternalServerError, "TO BE Implemented"
	default:
		log.Printf("[codesServerErrors] Unrecognized Status Code [%d]\n", code)
		return http.StatusInternalServerError, "Unknown Reason"
	}
}

// Convert a Code to a Message String
func CodeToMessage(code int) (int, string) {
	switch {
	case code >= 1000 && code < 2000:
		return codesOk(code)
	case code >= 2000 && code < 3000:
		return codesWarnings(code)
	case code >= 3000 && code < 4000:
		return codesProcessingErrors(code)
	case code >= 4000 && code < 5000:
		return codesObjectErrors(code)
	case code >= 5000 && code < 6000:
		return codesServerErrors(code)
	default:
		log.Printf("[codeToMessage] Unrecognized Status Code [%d]\n", code)
		return http.StatusServiceUnavailable, "Unknown Reason"
	}
}
