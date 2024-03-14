// Code generated by protoc-gen-go-web. DO NOT EDIT.
// version:
// 	protoc-gen-go-web v0.0.108<br />
// 	protoc            v4.25.3<br />
// source: api/im_chat.proto<br />

package api

import (
	context "context"
	cerror "github.com/chenjie199234/Corelib/cerror"
	log "github.com/chenjie199234/Corelib/log"
	metadata "github.com/chenjie199234/Corelib/metadata"
	web "github.com/chenjie199234/Corelib/web"
	protojson "google.golang.org/protobuf/encoding/protojson"
	proto "google.golang.org/protobuf/proto"
	io "io"
	http "net/http"
	strconv "strconv"
	strings "strings"
)

var _WebPathChatSend = "/im.chat/send"
var _WebPathChatRecall = "/im.chat/recall"
var _WebPathChatAck = "/im.chat/ack"
var _WebPathChatPull = "/im.chat/pull"

type ChatWebClient interface {
	// send a msg
	Send(context.Context, *SendReq, http.Header) (*SendResp, error)
	// recall a msg send by self
	Recall(context.Context, *RecallReq, http.Header) (*RecallResp, error)
	// already read the msg send by other(self's msg don't need to ack)
	Ack(context.Context, *AckReq, http.Header) (*AckResp, error)
	// get more msgs and recalls
	Pull(context.Context, *PullReq, http.Header) (*PullResp, error)
}

type chatWebClient struct {
	cc *web.WebClient
}

func NewChatWebClient(c *web.WebClient) ChatWebClient {
	return &chatWebClient{cc: c}
}

func (c *chatWebClient) Send(ctx context.Context, req *SendReq, header http.Header) (*SendResp, error) {
	if req == nil {
		return nil, cerror.ErrReq
	}
	if header == nil {
		header = make(http.Header)
	}
	header.Set("Content-Type", "application/x-protobuf")
	header.Set("Accept", "application/x-protobuf")
	reqd, _ := proto.Marshal(req)
	r, e := c.cc.Post(ctx, _WebPathChatSend, "", header, metadata.GetMetadata(ctx), reqd)
	if e != nil {
		return nil, e
	}
	data, e := io.ReadAll(r.Body)
	r.Body.Close()
	if e != nil {
		return nil, cerror.ConvertStdError(e)
	}
	resp := new(SendResp)
	if len(data) == 0 {
		return resp, nil
	}
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-protobuf") {
		if e := proto.Unmarshal(data, resp); e != nil {
			return nil, cerror.ErrResp
		}
	} else if e := (protojson.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}).Unmarshal(data, resp); e != nil {
		return nil, cerror.ErrResp
	}
	return resp, nil
}
func (c *chatWebClient) Recall(ctx context.Context, req *RecallReq, header http.Header) (*RecallResp, error) {
	if req == nil {
		return nil, cerror.ErrReq
	}
	if header == nil {
		header = make(http.Header)
	}
	header.Set("Content-Type", "application/x-protobuf")
	header.Set("Accept", "application/x-protobuf")
	reqd, _ := proto.Marshal(req)
	r, e := c.cc.Post(ctx, _WebPathChatRecall, "", header, metadata.GetMetadata(ctx), reqd)
	if e != nil {
		return nil, e
	}
	data, e := io.ReadAll(r.Body)
	r.Body.Close()
	if e != nil {
		return nil, cerror.ConvertStdError(e)
	}
	resp := new(RecallResp)
	if len(data) == 0 {
		return resp, nil
	}
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-protobuf") {
		if e := proto.Unmarshal(data, resp); e != nil {
			return nil, cerror.ErrResp
		}
	} else if e := (protojson.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}).Unmarshal(data, resp); e != nil {
		return nil, cerror.ErrResp
	}
	return resp, nil
}
func (c *chatWebClient) Ack(ctx context.Context, req *AckReq, header http.Header) (*AckResp, error) {
	if req == nil {
		return nil, cerror.ErrReq
	}
	if header == nil {
		header = make(http.Header)
	}
	header.Set("Content-Type", "application/x-protobuf")
	header.Set("Accept", "application/x-protobuf")
	reqd, _ := proto.Marshal(req)
	r, e := c.cc.Post(ctx, _WebPathChatAck, "", header, metadata.GetMetadata(ctx), reqd)
	if e != nil {
		return nil, e
	}
	data, e := io.ReadAll(r.Body)
	r.Body.Close()
	if e != nil {
		return nil, cerror.ConvertStdError(e)
	}
	resp := new(AckResp)
	if len(data) == 0 {
		return resp, nil
	}
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-protobuf") {
		if e := proto.Unmarshal(data, resp); e != nil {
			return nil, cerror.ErrResp
		}
	} else if e := (protojson.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}).Unmarshal(data, resp); e != nil {
		return nil, cerror.ErrResp
	}
	return resp, nil
}
func (c *chatWebClient) Pull(ctx context.Context, req *PullReq, header http.Header) (*PullResp, error) {
	if req == nil {
		return nil, cerror.ErrReq
	}
	if header == nil {
		header = make(http.Header)
	}
	header.Set("Content-Type", "application/x-protobuf")
	header.Set("Accept", "application/x-protobuf")
	reqd, _ := proto.Marshal(req)
	r, e := c.cc.Post(ctx, _WebPathChatPull, "", header, metadata.GetMetadata(ctx), reqd)
	if e != nil {
		return nil, e
	}
	data, e := io.ReadAll(r.Body)
	r.Body.Close()
	if e != nil {
		return nil, cerror.ConvertStdError(e)
	}
	resp := new(PullResp)
	if len(data) == 0 {
		return resp, nil
	}
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-protobuf") {
		if e := proto.Unmarshal(data, resp); e != nil {
			return nil, cerror.ErrResp
		}
	} else if e := (protojson.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}).Unmarshal(data, resp); e != nil {
		return nil, cerror.ErrResp
	}
	return resp, nil
}

type ChatWebServer interface {
	// send a msg
	Send(context.Context, *SendReq) (*SendResp, error)
	// recall a msg send by self
	Recall(context.Context, *RecallReq) (*RecallResp, error)
	// already read the msg send by other(self's msg don't need to ack)
	Ack(context.Context, *AckReq) (*AckResp, error)
	// get more msgs and recalls
	Pull(context.Context, *PullReq) (*PullResp, error)
}

func _Chat_Send_WebHandler(handler func(context.Context, *SendReq) (*SendResp, error)) web.OutsideHandler {
	return func(ctx *web.Context) {
		req := new(SendReq)
		if strings.HasPrefix(ctx.GetContentType(), "application/json") {
			data, e := ctx.GetBody()
			if e != nil {
				log.Error(ctx, "[/im.chat/send] get body failed", log.CError(e))
				ctx.Abort(e)
				return
			}
			if len(data) > 0 {
				if e := (protojson.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}).Unmarshal(data, req); e != nil {
					log.Error(ctx, "[/im.chat/send] unmarshal json body failed", log.CError(e))
					ctx.Abort(cerror.ErrReq)
					return
				}
			}
		} else if strings.HasPrefix(ctx.GetContentType(), "application/x-protobuf") {
			data, e := ctx.GetBody()
			if e != nil {
				log.Error(ctx, "[/im.chat/send] get body failed", log.CError(e))
				ctx.Abort(e)
				return
			}
			if len(data) > 0 {
				if e := proto.Unmarshal(data, req); e != nil {
					log.Error(ctx, "[/im.chat/send] unmarshal proto body failed", log.CError(e))
					ctx.Abort(cerror.ErrReq)
					return
				}
			}
		} else {
			if e := ctx.ParseForm(); e != nil {
				log.Error(ctx, "[/im.chat/send] parse form failed", log.CError(e))
				ctx.Abort(cerror.ErrReq)
				return
			}
			// req.Target
			if form := ctx.GetForm("target"); len(form) != 0 {
				req.Target = form
			}
			// req.TargetType
			if form := ctx.GetForm("target_type"); len(form) != 0 {
				req.TargetType = form
			}
			// req.Msg
			if form := ctx.GetForm("msg"); len(form) != 0 {
				req.Msg = form
			}
			// req.Extra
			if form := ctx.GetForm("extra"); len(form) != 0 {
				req.Extra = form
			}
		}
		if errstr := req.Validate(); errstr != "" {
			log.Error(ctx, "[/im.chat/send] validate failed", log.String("validate", errstr))
			ctx.Abort(cerror.ErrReq)
			return
		}
		resp, e := handler(ctx, req)
		ee := cerror.ConvertStdError(e)
		if ee != nil {
			ctx.Abort(ee)
			return
		}
		if resp == nil {
			resp = new(SendResp)
		}
		if strings.HasPrefix(ctx.GetAcceptType(), "application/x-protobuf") {
			respd, _ := proto.Marshal(resp)
			ctx.Write("application/x-protobuf", respd)
		} else {
			respd, _ := protojson.MarshalOptions{AllowPartial: true, UseProtoNames: true, UseEnumNumbers: true, EmitUnpopulated: true}.Marshal(resp)
			ctx.Write("application/json", respd)
		}
	}
}
func _Chat_Recall_WebHandler(handler func(context.Context, *RecallReq) (*RecallResp, error)) web.OutsideHandler {
	return func(ctx *web.Context) {
		req := new(RecallReq)
		if strings.HasPrefix(ctx.GetContentType(), "application/json") {
			data, e := ctx.GetBody()
			if e != nil {
				log.Error(ctx, "[/im.chat/recall] get body failed", log.CError(e))
				ctx.Abort(e)
				return
			}
			if len(data) > 0 {
				if e := (protojson.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}).Unmarshal(data, req); e != nil {
					log.Error(ctx, "[/im.chat/recall] unmarshal json body failed", log.CError(e))
					ctx.Abort(cerror.ErrReq)
					return
				}
			}
		} else if strings.HasPrefix(ctx.GetContentType(), "application/x-protobuf") {
			data, e := ctx.GetBody()
			if e != nil {
				log.Error(ctx, "[/im.chat/recall] get body failed", log.CError(e))
				ctx.Abort(e)
				return
			}
			if len(data) > 0 {
				if e := proto.Unmarshal(data, req); e != nil {
					log.Error(ctx, "[/im.chat/recall] unmarshal proto body failed", log.CError(e))
					ctx.Abort(cerror.ErrReq)
					return
				}
			}
		} else {
			if e := ctx.ParseForm(); e != nil {
				log.Error(ctx, "[/im.chat/recall] parse form failed", log.CError(e))
				ctx.Abort(cerror.ErrReq)
				return
			}
			// req.Target
			if form := ctx.GetForm("target"); len(form) != 0 {
				req.Target = form
			}
			// req.TargetType
			if form := ctx.GetForm("target_type"); len(form) != 0 {
				req.TargetType = form
			}
			// req.MsgIndex
			if form := ctx.GetForm("msg_index"); len(form) != 0 {
				if num, e := strconv.ParseUint(form, 10, 64); e != nil {
					log.Error(ctx, "[/im.chat/recall] data format wrong", log.String("field", "msg_index"))
					ctx.Abort(cerror.ErrReq)
					return
				} else {
					req.MsgIndex = num
				}
			}
		}
		if errstr := req.Validate(); errstr != "" {
			log.Error(ctx, "[/im.chat/recall] validate failed", log.String("validate", errstr))
			ctx.Abort(cerror.ErrReq)
			return
		}
		resp, e := handler(ctx, req)
		ee := cerror.ConvertStdError(e)
		if ee != nil {
			ctx.Abort(ee)
			return
		}
		if resp == nil {
			resp = new(RecallResp)
		}
		if strings.HasPrefix(ctx.GetAcceptType(), "application/x-protobuf") {
			respd, _ := proto.Marshal(resp)
			ctx.Write("application/x-protobuf", respd)
		} else {
			respd, _ := protojson.MarshalOptions{AllowPartial: true, UseProtoNames: true, UseEnumNumbers: true, EmitUnpopulated: true}.Marshal(resp)
			ctx.Write("application/json", respd)
		}
	}
}
func _Chat_Ack_WebHandler(handler func(context.Context, *AckReq) (*AckResp, error)) web.OutsideHandler {
	return func(ctx *web.Context) {
		req := new(AckReq)
		if strings.HasPrefix(ctx.GetContentType(), "application/json") {
			data, e := ctx.GetBody()
			if e != nil {
				log.Error(ctx, "[/im.chat/ack] get body failed", log.CError(e))
				ctx.Abort(e)
				return
			}
			if len(data) > 0 {
				if e := (protojson.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}).Unmarshal(data, req); e != nil {
					log.Error(ctx, "[/im.chat/ack] unmarshal json body failed", log.CError(e))
					ctx.Abort(cerror.ErrReq)
					return
				}
			}
		} else if strings.HasPrefix(ctx.GetContentType(), "application/x-protobuf") {
			data, e := ctx.GetBody()
			if e != nil {
				log.Error(ctx, "[/im.chat/ack] get body failed", log.CError(e))
				ctx.Abort(e)
				return
			}
			if len(data) > 0 {
				if e := proto.Unmarshal(data, req); e != nil {
					log.Error(ctx, "[/im.chat/ack] unmarshal proto body failed", log.CError(e))
					ctx.Abort(cerror.ErrReq)
					return
				}
			}
		} else {
			if e := ctx.ParseForm(); e != nil {
				log.Error(ctx, "[/im.chat/ack] parse form failed", log.CError(e))
				ctx.Abort(cerror.ErrReq)
				return
			}
			// req.Target
			if form := ctx.GetForm("target"); len(form) != 0 {
				req.Target = form
			}
			// req.TargetType
			if form := ctx.GetForm("target_type"); len(form) != 0 {
				req.TargetType = form
			}
			// req.MsgIndex
			if form := ctx.GetForm("msg_index"); len(form) != 0 {
				if num, e := strconv.ParseUint(form, 10, 64); e != nil {
					log.Error(ctx, "[/im.chat/ack] data format wrong", log.String("field", "msg_index"))
					ctx.Abort(cerror.ErrReq)
					return
				} else {
					req.MsgIndex = num
				}
			}
		}
		if errstr := req.Validate(); errstr != "" {
			log.Error(ctx, "[/im.chat/ack] validate failed", log.String("validate", errstr))
			ctx.Abort(cerror.ErrReq)
			return
		}
		resp, e := handler(ctx, req)
		ee := cerror.ConvertStdError(e)
		if ee != nil {
			ctx.Abort(ee)
			return
		}
		if resp == nil {
			resp = new(AckResp)
		}
		if strings.HasPrefix(ctx.GetAcceptType(), "application/x-protobuf") {
			respd, _ := proto.Marshal(resp)
			ctx.Write("application/x-protobuf", respd)
		} else {
			respd, _ := protojson.MarshalOptions{AllowPartial: true, UseProtoNames: true, UseEnumNumbers: true, EmitUnpopulated: true}.Marshal(resp)
			ctx.Write("application/json", respd)
		}
	}
}
func _Chat_Pull_WebHandler(handler func(context.Context, *PullReq) (*PullResp, error)) web.OutsideHandler {
	return func(ctx *web.Context) {
		req := new(PullReq)
		if strings.HasPrefix(ctx.GetContentType(), "application/json") {
			data, e := ctx.GetBody()
			if e != nil {
				log.Error(ctx, "[/im.chat/pull] get body failed", log.CError(e))
				ctx.Abort(e)
				return
			}
			if len(data) > 0 {
				if e := (protojson.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}).Unmarshal(data, req); e != nil {
					log.Error(ctx, "[/im.chat/pull] unmarshal json body failed", log.CError(e))
					ctx.Abort(cerror.ErrReq)
					return
				}
			}
		} else if strings.HasPrefix(ctx.GetContentType(), "application/x-protobuf") {
			data, e := ctx.GetBody()
			if e != nil {
				log.Error(ctx, "[/im.chat/pull] get body failed", log.CError(e))
				ctx.Abort(e)
				return
			}
			if len(data) > 0 {
				if e := proto.Unmarshal(data, req); e != nil {
					log.Error(ctx, "[/im.chat/pull] unmarshal proto body failed", log.CError(e))
					ctx.Abort(cerror.ErrReq)
					return
				}
			}
		} else {
			if e := ctx.ParseForm(); e != nil {
				log.Error(ctx, "[/im.chat/pull] parse form failed", log.CError(e))
				ctx.Abort(cerror.ErrReq)
				return
			}
			// req.Target
			if form := ctx.GetForm("target"); len(form) != 0 {
				req.Target = form
			}
			// req.TargetType
			if form := ctx.GetForm("target_type"); len(form) != 0 {
				req.TargetType = form
			}
			// req.Direction
			if form := ctx.GetForm("direction"); len(form) != 0 {
				req.Direction = form
			}
			// req.StartMsgIndex
			if form := ctx.GetForm("start_msg_index"); len(form) != 0 {
				if num, e := strconv.ParseUint(form, 10, 64); e != nil {
					log.Error(ctx, "[/im.chat/pull] data format wrong", log.String("field", "start_msg_index"))
					ctx.Abort(cerror.ErrReq)
					return
				} else {
					req.StartMsgIndex = num
				}
			}
			// req.StartRecallIndex
			if form := ctx.GetForm("start_recall_index"); len(form) != 0 {
				if num, e := strconv.ParseUint(form, 10, 64); e != nil {
					log.Error(ctx, "[/im.chat/pull] data format wrong", log.String("field", "start_recall_index"))
					ctx.Abort(cerror.ErrReq)
					return
				} else {
					req.StartRecallIndex = num
				}
			}
			// req.Count
			if form := ctx.GetForm("count"); len(form) != 0 {
				if num, e := strconv.ParseUint(form, 10, 64); e != nil {
					log.Error(ctx, "[/im.chat/pull] data format wrong", log.String("field", "count"))
					ctx.Abort(cerror.ErrReq)
					return
				} else {
					req.Count = num
				}
			}
		}
		if errstr := req.Validate(); errstr != "" {
			log.Error(ctx, "[/im.chat/pull] validate failed", log.String("validate", errstr))
			ctx.Abort(cerror.ErrReq)
			return
		}
		resp, e := handler(ctx, req)
		ee := cerror.ConvertStdError(e)
		if ee != nil {
			ctx.Abort(ee)
			return
		}
		if resp == nil {
			resp = new(PullResp)
		}
		if strings.HasPrefix(ctx.GetAcceptType(), "application/x-protobuf") {
			respd, _ := proto.Marshal(resp)
			ctx.Write("application/x-protobuf", respd)
		} else {
			respd, _ := protojson.MarshalOptions{AllowPartial: true, UseProtoNames: true, UseEnumNumbers: true, EmitUnpopulated: true}.Marshal(resp)
			ctx.Write("application/json", respd)
		}
	}
}
func RegisterChatWebServer(router *web.Router, svc ChatWebServer, allmids map[string]web.OutsideHandler) {
	// avoid lint
	_ = allmids
	{
		requiredMids := []string{"token"}
		mids := make([]web.OutsideHandler, 0, 2)
		for _, v := range requiredMids {
			if mid, ok := allmids[v]; ok {
				mids = append(mids, mid)
			} else {
				panic("missing midware:" + v)
			}
		}
		mids = append(mids, _Chat_Send_WebHandler(svc.Send))
		router.Post(_WebPathChatSend, mids...)
	}
	{
		requiredMids := []string{"token"}
		mids := make([]web.OutsideHandler, 0, 2)
		for _, v := range requiredMids {
			if mid, ok := allmids[v]; ok {
				mids = append(mids, mid)
			} else {
				panic("missing midware:" + v)
			}
		}
		mids = append(mids, _Chat_Recall_WebHandler(svc.Recall))
		router.Post(_WebPathChatRecall, mids...)
	}
	{
		requiredMids := []string{"token"}
		mids := make([]web.OutsideHandler, 0, 2)
		for _, v := range requiredMids {
			if mid, ok := allmids[v]; ok {
				mids = append(mids, mid)
			} else {
				panic("missing midware:" + v)
			}
		}
		mids = append(mids, _Chat_Ack_WebHandler(svc.Ack))
		router.Post(_WebPathChatAck, mids...)
	}
	{
		requiredMids := []string{"token"}
		mids := make([]web.OutsideHandler, 0, 2)
		for _, v := range requiredMids {
			if mid, ok := allmids[v]; ok {
				mids = append(mids, mid)
			} else {
				panic("missing midware:" + v)
			}
		}
		mids = append(mids, _Chat_Pull_WebHandler(svc.Pull))
		router.Post(_WebPathChatPull, mids...)
	}
}
