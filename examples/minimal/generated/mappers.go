package generated

import (
	"net/http"
	"github.com/xdrm-io/aicra/runtime"
)

type mapper struct {
	impl Server
}

func (m mapper) GetUsers(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		req GetUsersReq
	)

	res, err := m.impl.GetUsers(r.Context(), req)
	var out map[string]any
	if res != nil {
		out = map[string]any{
			"users": res.Users,
		}
	}
	runtime.Respond(w, out, err)
}

func (m mapper) GetUser(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		req GetUserReq
	)
	req.ID, err = runtime.ExtractURI[string](r, 1, getCustomUUIDValidator(nil))
	if err != nil {
		runtime.Respond(w, nil, err)
		return
	}

	res, err := m.impl.GetUser(r.Context(), req)
	var out map[string]any
	if res != nil {
		out = map[string]any{
			"firstname": res.Firstname,
			"lastname":  res.Lastname,
			"username":  res.Username,
		}
	}
	runtime.Respond(w, out, err)
}

func (m mapper) CreateUser(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		req CreateUserReq
	)
	form, err := runtime.ParseForm(r)
	if err != nil {
		runtime.Respond(w, nil, err)
		return
	}
	req.Firstname, err = runtime.ExtractForm[string](form, "firstname", getBuiltinStringValidator([]string{"1", "30"}))
	if err != nil {
		runtime.Respond(w, nil, err)
		return
	}
	req.Lastname, err = runtime.ExtractForm[string](form, "lastname", getBuiltinStringValidator([]string{"1", "30"}))
	if err != nil {
		runtime.Respond(w, nil, err)
		return
	}
	req.Username, err = runtime.ExtractForm[string](form, "username", getBuiltinStringValidator([]string{"1", "30"}))
	if err != nil {
		runtime.Respond(w, nil, err)
		return
	}

	res, err := m.impl.CreateUser(r.Context(), req)
	var out map[string]any
	if res != nil {
		out = map[string]any{
			"firstname": res.Firstname,
			"id":        res.ID,
			"lastname":  res.Lastname,
			"username":  res.Username,
		}
	}
	runtime.Respond(w, out, err)
}

func (m mapper) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		req UpdateUserReq
	)
	req.ID, err = runtime.ExtractURI[string](r, 1, getCustomUUIDValidator(nil))
	if err != nil {
		runtime.Respond(w, nil, err)
		return
	}
	form, err := runtime.ParseForm(r)
	if err != nil {
		runtime.Respond(w, nil, err)
		return
	}
	InFirstname, err := runtime.ExtractForm[string](form, "firstname", getBuiltinStringValidator(nil))
	if err != nil && err != runtime.ErrMissingParam {
		runtime.Respond(w, nil, err)
		return
	}
	if err == nil {
		req.Firstname = &InFirstname
	}
	InLastname, err := runtime.ExtractForm[string](form, "lastname", getBuiltinStringValidator(nil))
	if err != nil && err != runtime.ErrMissingParam {
		runtime.Respond(w, nil, err)
		return
	}
	if err == nil {
		req.Lastname = &InLastname
	}
	InUsername, err := runtime.ExtractForm[string](form, "username", getBuiltinStringValidator(nil))
	if err != nil && err != runtime.ErrMissingParam {
		runtime.Respond(w, nil, err)
		return
	}
	if err == nil {
		req.Username = &InUsername
	}

	res, err := m.impl.UpdateUser(r.Context(), req)
	var out map[string]any
	if res != nil {
		out = map[string]any{
			"firstname": res.Firstname,
			"id":        res.ID,
			"lastname":  res.Lastname,
			"username":  res.Username,
		}
	}
	runtime.Respond(w, out, err)
}

func (m mapper) DeleteUser(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		req DeleteUserReq
	)
	req.ID, err = runtime.ExtractURI[string](r, 1, getCustomUUIDValidator(nil))
	if err != nil {
		runtime.Respond(w, nil, err)
		return
	}
	_, err = m.impl.DeleteUser(r.Context(), req)
	runtime.Respond(w, nil, err)
}
