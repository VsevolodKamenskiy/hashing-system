package api

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"service2/internal/grpcclient"
	"service2/internal/mw"
	"service2/internal/storage"
)

type Handlers struct {
	HashClient grpcclient.HasherClient
	Store      *storage.Store
	Log        *logrus.Logger
}

// POST /send
// body: ["str1","str2",...]
// 200: [{"id":38,"hash":"..."}]
func (h *Handlers) Send(c *gin.Context) {
	var in []string
	if err := c.ShouldBindJSON(&in); err != nil {
		werr := errors.WithStack(err)
		h.Log.WithField("stack", fmt.Sprintf("%+v", werr)).WithError(werr).
			Error("send: bad request")
		c.Status(http.StatusBadRequest)
		return
	}
	if len(in) == 0 {
		c.JSON(http.StatusOK, []any{})
		return
	}

	reqID := mw.FromContext(c.Request.Context())
	h.Log.WithField("request_id", reqID).WithField("count", len(in)).Info("send: hashing")

	hashes, err := h.HashClient.Calculate(c.Request.Context(), in)
	if err != nil {
		werr := errors.WithStack(err)
		h.Log.WithField("request_id", mw.FromContext(c.Request.Context())).
			WithField("stack", fmt.Sprintf("%+v", werr)).WithError(werr).
			Error("send: grpc call failed")
		c.Status(http.StatusInternalServerError)
		return
	}

	rows, err := h.Store.InsertHashes(c.Request.Context(), hashes)
	if err != nil {
		werr := errors.WithStack(err)
		h.Log.WithField("request_id", mw.FromContext(c.Request.Context())).
			WithField("stack", fmt.Sprintf("%+v", werr)).WithError(werr).
			Error("send: db insert failed")
		c.Status(http.StatusInternalServerError)
		return
	}

	type resp struct {
		ID   int64  `json:"id"`
		Hash string `json:"hash"`
	}
	out := make([]resp, 0, len(rows))
	for _, r := range rows {
		out = append(out, resp{ID: r.ID, Hash: r.Hash})
	}

	h.Log.WithField("request_id", reqID).WithField("saved", len(out)).Info("send: done")
	c.JSON(http.StatusOK, out)
}

// GET /check?ids=1&ids=2 или /check?ids=1,2
// 200: [{"id":38,"hash":"..."}], 204 если нет совпадений
func (h *Handlers) Check(c *gin.Context) {
	idsParam := c.QueryArray("ids")
	if len(idsParam) == 0 {
		if raw := c.Query("ids"); raw != "" {
			idsParam = strings.Split(raw, ",")
		}
	}
	if len(idsParam) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	var ids []int64
	for _, s := range idsParam {
		if s == "" {
			continue
		}
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		ids = append(ids, v)
	}

	reqID := mw.FromContext(c.Request.Context())
	h.Log.WithField("request_id", reqID).WithField("ids", ids).Info("check: start")

	rows, err := h.Store.GetByIDs(c.Request.Context(), ids)
	if err != nil {
		werr := errors.WithStack(err)
		h.Log.WithField("request_id", mw.FromContext(c.Request.Context())).
			WithField("stack", fmt.Sprintf("%+v", werr)).WithError(werr).
			Error("check: db failed")
		c.Status(http.StatusInternalServerError)
		return
	}
	if len(rows) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	type resp struct {
		ID   int64  `json:"id"`
		Hash string `json:"hash"`
	}
	out := make([]resp, 0, len(rows))
	for _, r := range rows {
		out = append(out, resp{ID: r.ID, Hash: r.Hash})
	}
	h.Log.WithField("request_id", reqID).WithField("found", len(out)).Info("check: done")
	c.JSON(http.StatusOK, out)
}
