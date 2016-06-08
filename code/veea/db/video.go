package db

import (
	"time"
	"errors"
	"sync"
"github.com/gpahal/veea/conf"
)

type Video struct {
	VideoId   string
	Name      string
	CreatedAt time.Time
}

type ViewTime struct {
	ViewId  string
	Time    float64
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

var (
	viewIdLock sync.Mutex
)

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

func AddVideoView(userId int64, videoId string) (string, error) {
	err := errorFold(
		UserIdExists(userId),
		VideoIdExists(videoId),
	)
	if err != nil {
		return "", &UserError{error: err}
	}

	viewIdLock.Lock()
	defer viewIdLock.Unlock()

	viewId, err := GenerateViewId()
	if err != nil {
		return "", &InternalError{error: err}
	}

	res, err := exec("INSERT INTO video_view (user_id, video_id, view_id) VALUES (?, ?, ?)", userId, videoId, viewId)
	if err != nil {
		return "", &InternalError{error: err}
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return "", &InternalError{error: err}
	}
	if ra < 1 {
		return "", &InternalError{error: errors.New("Add video view failed")}
	}

	return viewId, nil
}

func AddViewTime(videoId string, viewTime *ViewTime) (int64, error) {
	err := errorFold(
		ViewIdNotExpiredExists(videoId, viewTime.ViewId),
	)
	if err != nil {
		return 0, &UserError{error: err}
	}

	res, err := exec("INSERT INTO video_view_time (view_id, time, state, quality) VALUES (?, ?, ?, ?)", viewTime.ViewId, viewTime.Time, viewTime.State, viewTime.Quality)
	if err != nil {
		return 0, &InternalError{error: err}
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, &InternalError{error: err}
	}

	return id, nil
}

func AddViewStats(viewStats *ViewStats) (int64, error) {
	err := errorFold(
		ViewTimeIdExists(viewStats.ViewTimeId),
	)
	if err != nil {
		return 0, &UserError{error: err}
	}

	columns := "view_time_id, gender, age, mood, head_yaw, head_pitch, head_roll, head_x, head_y, head_z, head_gaze_x, head_gaze_y, happy, surprised, angry, disgusted, afraid, sad, engagement"

	res, err := exec("INSERT INTO video_view_stats (" + columns + ") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		viewStats.ViewTimeId,
		viewStats.Gender,
		viewStats.Age,
		viewStats.Mood,
		viewStats.HeadYaw,
		viewStats.HeadPitch,
		viewStats.HeadRoll,
		viewStats.HeadX,
		viewStats.HeadY,
		viewStats.HeadZ,
		viewStats.HeadGazeX,
		viewStats.HeadGazeY,
		viewStats.Happy,
		viewStats.Surprised,
		viewStats.Angry,
		viewStats.Disgusted,
		viewStats.Afraid,
		viewStats.Sad,
		viewStats.Engagement,
	)
	if err != nil {
		return 0, &InternalError{error: err}
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, &InternalError{error: err}
	}

	return id, nil
}

func UpdateVideoDuration() error {
	currentTime := time.Now().Unix()
	maxExpireTime := time.Unix(currentTime - conf.ViewExpireTime, 0)

	rows, err := query("SELECT view_id FROM video_view WHERE view_duration >= 0 AND created_at < ?", maxExpireTime)
	if err != nil {
		return &InternalError{error: err}
	}
	defer rows.Close()

	for rows.Next() {
		var videoId string
		err = rows.Scan(&videoId)

		if err != nil {
			return &InternalError{error: err}
		}

		err = UpdateVideoDurationSingle(videoId)
		if err != nil {
			return &InternalError{error: err}
		}
	}

	return nil
}

func UpdateVideoDurationSingle(viewId string) error {
	rows, err := query("SELECT time FROM video_view_time WHERE view_id = ?", viewId)
	if err != nil {
		return &InternalError{error: err}
	}
	defer rows.Close()

	times := []int{}

	for rows.Next() {
		var time float64
		err = rows.Scan(&time)

		if err != nil {
			return &InternalError{error: err}
		}

		index := int(time / 5)
		currentLength := len(times)

		if index < currentLength {
			times[index] += 1
		} else {
			incrementRequired := index - currentLength + 1
			for i := 0; i < incrementRequired; i += 1 {
				times = append(times, 0)
			}
			times[index] += 1
		}
	}

	count := 0
	for _, timeSingle := range times {
		if timeSingle > 0 {
			count += 1
		}
	}

	_, err = exec("UPDATE video_view SET view_duration = ? WHERE view_id = ?", float64(count * 5), viewId)
	if err != nil {
		return &InternalError{error: err}
	}

	return nil
}