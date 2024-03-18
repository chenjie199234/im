// Code generated by protoc-gen-go-pbex. DO NOT EDIT.
// version:
// 	protoc-gen-pbex v0.0.109
// 	protoc         v4.25.3
// source: api/im_chat.proto

package api

// return empty means pass
func (m *SendReq) Validate() (errstr string) {
	if len(m.GetTarget()) <= 0 {
		return "field: target in object: send_req check value str len gt failed"
	}
	if m.GetTargetType() != "user" && m.GetTargetType() != "group" {
		return "field: target_type in object: send_req check value str in failed"
	}
	if len(m.GetMsg()) <= 0 {
		return "field: msg in object: send_req check value str len gt failed"
	}
	return ""
}

// return empty means pass
func (m *RecallReq) Validate() (errstr string) {
	if len(m.GetTarget()) <= 0 {
		return "field: target in object: recall_req check value str len gt failed"
	}
	if m.GetTargetType() != "user" && m.GetTargetType() != "group" {
		return "field: target_type in object: recall_req check value str in failed"
	}
	return ""
}

// return empty means pass
func (m *AckReq) Validate() (errstr string) {
	if len(m.GetTarget()) <= 0 {
		return "field: target in object: ack_req check value str len gt failed"
	}
	if m.GetTargetType() != "user" && m.GetTargetType() != "group" {
		return "field: target_type in object: ack_req check value str in failed"
	}
	return ""
}

// return empty means pass
func (m *PullReq) Validate() (errstr string) {
	if len(m.GetTarget()) <= 0 {
		return "field: target in object: pull_req check value str len gt failed"
	}
	if m.GetTargetType() != "user" && m.GetTargetType() != "group" {
		return "field: target_type in object: pull_req check value str in failed"
	}
	if m.GetDirection() != "before" && m.GetDirection() != "after" {
		return "field: direction in object: pull_req check value str in failed"
	}
	if m.GetCount() <= 0 {
		return "field: count in object: pull_req check value uint gt failed"
	}
	if m.GetCount() > 50 {
		return "field: count in object: pull_req check value uint lte failed"
	}
	return ""
}
