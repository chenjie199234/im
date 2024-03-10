package ecode

import (
	"net/http"

	"github.com/chenjie199234/Corelib/cerror"
)

var (
	ErrServerClosing     = cerror.ErrServerClosing     //1000  // http code 449 Warning!! Client will retry on this error,be careful to use this error
	ErrDataConflict      = cerror.ErrDataConflict      //9001  // http code 500
	ErrDataBroken        = cerror.ErrDataBroken        //9002  // http code 500
	ErrDBDataConflict    = cerror.ErrDBDataConflict    //9101  // http code 500
	ErrDBDataBroken      = cerror.ErrDBDataBroken      //9102  // http code 500
	ErrCacheDataConflict = cerror.ErrCacheDataConflict //9201  // http code 500
	ErrCacheDataBroken   = cerror.ErrCacheDataBroken   //9202  // http code 500
	ErrMQDataBroken      = cerror.ErrMQDataBroken      //9301  // http code 500
	ErrUnknown           = cerror.ErrUnknown           //10000 // http code 500
	ErrReq               = cerror.ErrReq               //10001 // http code 400
	ErrResp              = cerror.ErrResp              //10002 // http code 500
	ErrSystem            = cerror.ErrSystem            //10003 // http code 500
	ErrToken             = cerror.ErrToken             //10004 // http code 401
	ErrSession           = cerror.ErrSession           //10005 // http code 401
	ErrAccessKey         = cerror.ErrAccessKey         //10006 // http code 401
	ErrAccessSign        = cerror.ErrAccessSign        //10007 // http code 401
	ErrPermission        = cerror.ErrPermission        //10008 // http code 403
	ErrTooFast           = cerror.ErrTooFast           //10009 // http code 403
	ErrBan               = cerror.ErrBan               //10010 // http code 403
	ErrBusy              = cerror.ErrBusy              //10011 // http code 503
	ErrNotExist          = cerror.ErrNotExist          //10012 // http code 404
	ErrPasswordWrong     = cerror.ErrPasswordWrong     //10013 // http code 400
	ErrPasswordLength    = cerror.ErrPasswordLength    //10014 // http code 400

	ErrMsgNotExist   = cerror.MakeError(20001, http.StatusBadRequest, "msg not exist")
	ErrMsgOwnerWrong = cerror.MakeError(20002, http.StatusBadRequest, "msg owner wrong")

	ErrRequestNotExist = cerror.MakeError(21001, http.StatusBadRequest, "request not exist")
	ErrTooPopular      = cerror.MakeError(21002, http.StatusBadRequest, "target is too popular to get more requests")

	ErrUserNotExist  = cerror.MakeError(22001, http.StatusBadRequest, "user not exist")
	ErrGroupNotExist = cerror.MakeError(22002, http.StatusBadRequest, "group not exist")
	//in group's view,use below
	ErrGroupMemberNotExist = cerror.MakeError(22003, http.StatusBadRequest, "group member not exist")
	//in user's view,use below
	ErrNotFriends             = cerror.MakeError(22004, http.StatusBadRequest, "you are not friends")
	ErrNotInGroup             = cerror.MakeError(22005, http.StatusBadRequest, "you are not in the group")
	ErrSelfTooManyRelations   = cerror.MakeError(22006, http.StatusBadRequest, "you have too many friends and groups to get more")
	ErrTargetTooManyRelations = cerror.MakeError(22007, http.StatusBadRequest, "target has too many friends and groups to get more")
	ErrGroupTooManyMembers    = cerror.MakeError(22008, http.StatusBadRequest, "group has too many members to get more")
)

func ReturnEcode(originerror error, defaulterror *cerror.Error) error {
	if _, ok := originerror.(*cerror.Error); ok {
		return originerror
	}
	return defaulterror
}
