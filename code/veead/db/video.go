package db

import (
	"time"
	"errors"
)

type Video struct {
	VideoId   string
	Name      string
	CreatedAt time.Time
}

type View struct {
	UserId int64
	VideoId string
	ViewId string
	VideoDuration float64
	CreatedAt time.Time
}

type ViewTime struct {
	ViewId  string
	Ratio   float64
	State   int
	Quality string
}

type ViewStats struct {
	ViewTimeId int64
	Gender     float64
	Age        int
	Mood       float64
	HeadYaw    float64
	HeadPitch  float64
	HeadRoll   float64
	HeadX      float64
	HeadY      float64
	HeadZ      float64
	HeadGazeX  float64
	HeadGazeY  float64
	Happy      float64
	Surprised  float64
	Angry      float64
	Disgusted  float64
	Afraid     float64
	Sad        float64
	Engagement float64
}

func GetVideos(userId int64) ([]*Video, error) {
	err := errorFold(
		UserIdAdminExists(userId),
	)
	if err != nil {
		return nil, &UserError{error: err}
	}

	rows, err := query("SELECT video_id, name, created_at FROM video ORDER BY created_at DESC")
	if err != nil {
		return nil, &InternalError{error: err}
	}
	defer rows.Close()

	videos := []*Video{}

	for rows.Next() {
		var video Video
		err = rows.Scan(
			&video.VideoId,
			&video.Name,
			&video.CreatedAt,
		)

		if err != nil {
			return nil, &InternalError{error: err}
		}

		videos = append(videos, &video)
	}

	return videos, nil
}

func GetVideo(videoId string) (*Video, error) {
	rows, err := query("SELECT video_id, name, created_at FROM video WHERE video_id = ? LIMIT 1", videoId)
	if err != nil {
		return nil, &InternalError{error: err}
	}
	defer rows.Close()

	if rows.Next() {
		var video Video
		err = rows.Scan(
			&video.VideoId,
			&video.Name,
			&video.CreatedAt,
		)

		if err != nil {
			return nil, &InternalError{error: err}
		}

		return &video, nil
	}

	return nil, &UserError{error: errors.New("Video does not exist")}
}

func GetViews(userId int64, otherUserId int64) ([]*View, error) {
	err := errorFold(
		UserIdAdminExists(userId),
	)
	if err != nil {
		return nil, &UserError{error: err}
	}

	rows, err := query("SELECT user_id, video_id, view_id, view_duration, created_at FROM video_view WHERE user_id = ? ORDER BY created_at DESC", otherUserId)
	if err != nil {
		return nil, &InternalError{error: err}
	}
	defer rows.Close()

	views := []*View{}

	for rows.Next() {
		var view View
		err = rows.Scan(
			&view.UserId,
			&view.VideoId,
			&view.ViewId,
			&view.VideoDuration,
			&view.CreatedAt,
		)

		if err != nil {
			return nil, &InternalError{error: err}
		}

		views = append(views, &view)
	}

	return views, nil
}

func AddVideo(userId int64, videoId string, name string) error {
	err := errorFold(
		UserIdAdminExists(userId),
		VideoIdNotExists(videoId),
	)
	if err != nil {
		return &UserError{error: err}
	}

	_, err = exec("INSERT INTO video (video_id, name) VALUES (?, ?)", videoId, name)
	if err != nil {
		return &InternalError{error: err}
	}

	return nil
}

func UpdateVideo(userId int64, videoId string, name string) error {
	err := errorFold(
		UserIdAdminExists(userId),
	)
	if err != nil {
		return &UserError{error: err}
	}

	_, err = exec("UPDATE video SET name = ? WHERE video_id = ?", name, videoId)
	if err != nil {
		return &InternalError{error: err}
	}

	return nil
}

func DeleteVideo(userId int64, videoId string) error {
	err := errorFold(
		UserIdAdminExists(userId),
	)
	if err != nil {
		return &UserError{error: err}
	}

	_, err = exec("DELETE FROM video WHERE video_id = ?", videoId)
	if err != nil {
		return &InternalError{error: err}
	}

	return nil
}