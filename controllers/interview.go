package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/GetStream/getstream-go"
	"github.com/danilovict2/go-interview-RTC/models"
	uuidUtils "github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func (cfg *APIConfig) InterviewStore(c echo.Context) error {
	user, ok := c.Get("authUser").(models.User)
	if !ok {
		return HandleGracefully(fmt.Errorf("failed to retrieve authenticated user from context"), c)
	}

	if user.Role != models.ROLE_INTERVIEWER {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error": "Candidates can not create interviews",
		})
	}

	startTime, err := time.Parse(http.TimeFormat, c.FormValue("startTime"))
	if err != nil {
		return HandleGracefully(err, c)
	}

	status := models.STATUS_LIVE
	if startTime.After(time.Now()) {
		status = models.STATUS_UPCOMING
	}

	callID := uuidUtils.NewString()
	call := cfg.StreamClient.Video().Call("default", callID)
	callRequest := getstream.GetOrCreateCallRequest{
		Data: &getstream.CallRequest{
			CreatedByID: getstream.PtrTo(user.UUID.String()),
		},
	}

	resp, err := call.GetOrCreate(c.Request().Context(), &callRequest)
	if err != nil {
		return HandleGracefully(err, c)
	}

	interview := models.Interview{
		Title:        c.FormValue("title"),
		Description:  c.FormValue("description"),
		StartTime:    startTime,
		Status:       status,
		StreamCallID: resp.Data.Call.ID,
		Attendees:    []models.User{user},
	}

	if err := cfg.DB.Create(&interview).Error; err != nil {
		return HandleGracefully(err, c)
	}

	return c.JSON(http.StatusOK, interview)
}

func (cfg *APIConfig) InterviewEnd(c echo.Context) error {
	streamCallID := c.Param("stream-call-id")
	interview := models.Interview{}
	err := cfg.DB.First(&interview, "stream_call_id = ?", streamCallID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(http.StatusNotFound, echo.Map{
			"error": "Interview not found",
		})
	} else if err != nil {
		return HandleGracefully(err, c)
	}

	user, ok := c.Get("authUser").(models.User)
	if !ok {
		return HandleGracefully(fmt.Errorf("failed to retrieve authenticated user from context"), c)
	}

	attendees := make([]models.User, 0)
	if err := cfg.DB.Model(&interview).Association("Attendees").Find(&attendees); err != nil {
		return HandleGracefully(err, c)
	}

	isAttendee := false
	for _, attendee := range attendees {
		if attendee.UUID == user.UUID {
			isAttendee = true
			break
		}
	}

	if user.Role != models.ROLE_INTERVIEWER || !isAttendee {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error": "You can not modify this resource",
		})
	}

	endTime, err := time.Parse(http.TimeFormat, c.FormValue("endTime"))
	if err != nil {
		return HandleGracefully(err, c)
	}

	interview.Status = models.STATUS_COMPLETED
	interview.EndTime = &endTime

	if err := cfg.DB.Save(interview).Error; err != nil {
		return HandleGracefully(err, c)
	}

	return c.JSON(http.StatusOK, interview)
}
