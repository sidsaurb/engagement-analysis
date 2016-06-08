package db

import (
	"errors"
	"database/sql"
)

func GetTotalViews(videoId string) (int64, error) {
	rows, err := query("SELECT COUNT(*) FROM video_view WHERE video_id = ?", videoId)
	if err != nil {
		return 0, &InternalError{error: err}
	}
	defer rows.Close()

	if rows.Next() {
		var count sql.NullInt64
		err = rows.Scan(&count)

		if err != nil {
			return 0, &InternalError{error: err}
		}

		if count.Valid {
			return count.Int64, nil
		} else {
			return 0, nil
		}
	}

	return 0, errors.New("COUNT(*) returned 0 rows")
}

func GetUniqueVisitors(videoId string) (int64, error) {
	rows, err := query("SELECT COUNT(DISTINCT user_id) FROM video_view WHERE video_id = ?", videoId)
	if err != nil {
		return 0, &InternalError{error: err}
	}
	defer rows.Close()

	if rows.Next() {
		var count sql.NullInt64
		err = rows.Scan(&count)

		if err != nil {
			return 0, &InternalError{error: err}
		}

		if count.Valid {
			return count.Int64, nil
		} else {
			return 0, nil
		}
	}

	return 0, errors.New("COUNT(*) returned 0 rows")
}

func GetAverageViewDuration(videoId string, videoDuration float64) (float64, bool, error) {
	rows, err := query("SELECT AVG(view_duration) FROM video_view WHERE video_id = ? AND view_duration >= 0 AND view_duration <= ?", videoId, videoDuration)
	if err != nil {
		return 0, false, &InternalError{error: err}
	}
	defer rows.Close()

	if rows.Next() {
		var duration sql.NullFloat64
		err = rows.Scan(&duration)

		if err != nil {
			return 0, false, &InternalError{error: err}
		}

		if duration.Valid {
			return duration.Float64, true, nil
		} else {
			return 0, false, nil
		}
	}

	return 0, false, errors.New("Average view duration query returned 0 rows")
}

func GetMaleCount(videoId string) (int64, error) {
	rows, err := query("SELECT COUNT(*) FROM video_view_stats WHERE view_time_id IN (SELECT B.id FROM video_view AS A, video_view_time AS B WHERE A.video_id = ? AND A.view_id = B.view_id) AND gender < 0", videoId)
	if err != nil {
		return 0, &InternalError{error: err}
	}
	defer rows.Close()

	if rows.Next() {
		var count sql.NullInt64
		err = rows.Scan(&count)

		if err != nil {
			return 0, &InternalError{error: err}
		}

		if count.Valid {
			return count.Int64, nil
		} else {
			return 0, nil
		}
	}

	return 0, errors.New("COUNT(*) returned 0 rows")
}

func GetFemaleCount(videoId string) (int64, error) {
	rows, err := query("SELECT COUNT(*) FROM video_view_stats WHERE view_time_id IN (SELECT B.id FROM video_view AS A, video_view_time AS B WHERE A.video_id = ? AND A.view_id = B.view_id) AND gender > 0", videoId)
	if err != nil {
		return 0, &InternalError{error: err}
	}
	defer rows.Close()

	if rows.Next() {
		var count sql.NullInt64
		err = rows.Scan(&count)

		if err != nil {
			return 0, &InternalError{error: err}
		}

		if count.Valid {
			return count.Int64, nil
		} else {
			return 0, nil
		}
	}

	return 0, errors.New("COUNT(*) returned 0 rows")
}

func GetAgeCounts(videoId string) ([]int64, error) {
	rows, err := query("SELECT age, COUNT(*) FROM video_view_stats WHERE view_time_id IN (SELECT B.id FROM video_view AS A, video_view_time AS B WHERE A.video_id = ? AND A.view_id = B.view_id) GROUP BY age", videoId)
	if err != nil {
		return nil, &InternalError{error: err}
	}
	defer rows.Close()

	// age ranges: <18, 28-30, 31-50, >50
	ageCounts := make([]int64, 4, 4)

	for rows.Next() {
		var age int
		var count sql.NullInt64
		err = rows.Scan(&age, &count)

		if err != nil {
			return nil, &InternalError{error: err}
		}

		if count.Valid {
			if age < 0 {
				continue
			} else if age < 18 {
				ageCounts[0] += 1
			} else if age < 31 {
				ageCounts[1] += 1
			} else if age < 51 {
				ageCounts[2] += 1
			} else {
				ageCounts[3] += 1
			}
		}
	}

	return ageCounts, nil
}

func GetMaxEngagement() (float64, error) {
	rows, err := query("SELECT MAX(engagement) FROM video_view_stats")
	if err != nil {
		return 0, &InternalError{error: err}
	}
	defer rows.Close()

	if rows.Next() {
		var engagement float64
		err = rows.Scan(&engagement)

		if err != nil {
			return 0, &InternalError{error: err}
		}

		return engagement, nil
	}

	return 0, errors.New("Max engagement query returned 0 rows")
}

func GetStats(videoId string) ([]float64, error) {
	rows, err := query("SELECT AVG(mood), AVG(happy), AVG(surprised), AVG(angry), AVG(disgusted), AVG(afraid), AVG(sad), AVG(engagement) FROM video_view_stats WHERE view_time_id IN (SELECT B.id FROM video_view AS A, video_view_time AS B WHERE A.video_id = ? AND A.view_id = B.view_id)", videoId)
	if err != nil {
		return nil, &InternalError{error: err}
	}
	defer rows.Close()

	emotionValues := make([]float64, 8, 8)
	emotionSqlValues := make([]sql.NullFloat64, 8, 8)

	if rows.Next() {
		err = rows.Scan(
			&emotionSqlValues[0],
			&emotionSqlValues[1],
			&emotionSqlValues[2],
			&emotionSqlValues[3],
			&emotionSqlValues[4],
			&emotionSqlValues[5],
			&emotionSqlValues[6],
			&emotionSqlValues[7],
		)

		if err != nil {
			return nil, &InternalError{error: err}
		}

		for idx, value := range emotionSqlValues {
			if value.Valid {
				emotionValues[idx] = value.Float64
			} else {
				if idx == 7 {
					emotionValues[idx] = 1
				} else {
					emotionValues[idx] = 0
				}
			}
		}

		return emotionValues, nil
	}

	return nil, errors.New("Mood and emotions query returned 0 rows")
}

func GetInstantStats(videoId string, startTime float64, endTime float64) ([]float64, error) {
	rows, err := query("SELECT AVG(mood), AVG(happy), AVG(surprised), AVG(angry), AVG(disgusted), AVG(afraid), AVG(sad), AVG(engagement) FROM video_view_stats WHERE view_time_id IN (SELECT B.id FROM video_view AS A, video_view_time AS B WHERE A.video_id = ? AND A.view_id = B.view_id AND B.time >= ? AND B.time < ?)", videoId, startTime, endTime)
	if err != nil {
		return nil, &InternalError{error: err}
	}
	defer rows.Close()

	emotionValues := make([]float64, 8, 8)
	emotionSqlValues := make([]sql.NullFloat64, 8, 8)

	if rows.Next() {
		err = rows.Scan(
			&emotionSqlValues[0],
			&emotionSqlValues[1],
			&emotionSqlValues[2],
			&emotionSqlValues[3],
			&emotionSqlValues[4],
			&emotionSqlValues[5],
			&emotionSqlValues[6],
			&emotionSqlValues[7],
		)

		if err != nil {
			return nil, &InternalError{error: err}
		}

		for idx, value := range emotionSqlValues {
			if value.Valid {
				emotionValues[idx] = value.Float64
			} else {
				if idx == 7 {
					emotionValues[idx] = 1
				} else {
					emotionValues[idx] = 0
				}
			}
		}

		return emotionValues, nil
	}

	return nil, errors.New("Mood and emotions instant query returned 0 rows")
}

func GetInstantViewedCount(videoId string, startTime float64, endTime float64) (int64, error) {
	rows, err := query("SELECT COUNT(DISTINCT B.id) FROM video_view AS A, video_view_time AS B WHERE A.video_id = ? AND A.view_id = B.view_id AND B.time >= ? AND B.time < ?", videoId, startTime, endTime)
	if err != nil {
		return 0, &InternalError{error: err}
	}
	defer rows.Close()

	if rows.Next() {
		var count sql.NullInt64
		err = rows.Scan(&count)

		if err != nil {
			return 0, &InternalError{error: err}
		}

		if count.Valid {
			return count.Int64, nil
		} else {
			return 0, nil
		}
	}

	return 0, errors.New("COUNT(*) returned 0 rows")
}

func GetAverageViewDurationSingle(viewId string, videoDuration float64) (float64, bool, error) {
	rows, err := query("SELECT AVG(view_duration) FROM video_view WHERE view_id = ? AND view_duration >= 0 AND view_duration <= ?", viewId, videoDuration)
	if err != nil {
		return 0, false, &InternalError{error: err}
	}
	defer rows.Close()

	if rows.Next() {
		var duration sql.NullFloat64
		err = rows.Scan(&duration)

		if err != nil {
			return 0, false, &InternalError{error: err}
		}

		if duration.Valid {
			return duration.Float64, true, nil
		} else {
			return 0, false, nil
		}
	}

	return 0, false, errors.New("Average view duration query returned 0 rows")
}

func GetMaleCountSingle(viewId string) (int64, error) {
	rows, err := query("SELECT COUNT(*) FROM video_view_stats WHERE view_time_id IN (SELECT B.id FROM video_view AS A, video_view_time AS B WHERE A.view_id = ? AND A.view_id = B.view_id) AND gender < 0", viewId)
	if err != nil {
		return 0, &InternalError{error: err}
	}
	defer rows.Close()

	if rows.Next() {
		var count sql.NullInt64
		err = rows.Scan(&count)

		if err != nil {
			return 0, &InternalError{error: err}
		}

		if count.Valid {
			return count.Int64, nil
		} else {
			return 0, nil
		}
	}

	return 0, errors.New("COUNT(*) returned 0 rows")
}

func GetFemaleCountSingle(viewId string) (int64, error) {
	rows, err := query("SELECT COUNT(*) FROM video_view_stats WHERE view_time_id IN (SELECT B.id FROM video_view AS A, video_view_time AS B WHERE A.view_id = ? AND A.view_id = B.view_id) AND gender > 0", viewId)
	if err != nil {
		return 0, &InternalError{error: err}
	}
	defer rows.Close()

	if rows.Next() {
		var count sql.NullInt64
		err = rows.Scan(&count)

		if err != nil {
			return 0, &InternalError{error: err}
		}

		if count.Valid {
			return count.Int64, nil
		} else {
			return 0, nil
		}
	}

	return 0, errors.New("COUNT(*) returned 0 rows")
}

func GetAgeCountsSingle(viewId string) ([]int64, error) {
	rows, err := query("SELECT age, COUNT(*) FROM video_view_stats WHERE view_time_id IN (SELECT B.id FROM video_view AS A, video_view_time AS B WHERE A.view_id = ? AND A.view_id = B.view_id) GROUP BY age", viewId)
	if err != nil {
		return nil, &InternalError{error: err}
	}
	defer rows.Close()

	// age ranges: <18, 28-30, 31-50, >50
	ageCounts := make([]int64, 4, 4)

	for rows.Next() {
		var age int
		var count sql.NullInt64
		err = rows.Scan(&age, &count)

		if err != nil {
			return nil, &InternalError{error: err}
		}

		if count.Valid {
			if age < 0 {
				continue
			} else if age < 18 {
				ageCounts[0] += 1
			} else if age < 31 {
				ageCounts[1] += 1
			} else if age < 51 {
				ageCounts[2] += 1
			} else {
				ageCounts[3] += 1
			}
		}
	}

	return ageCounts, nil
}

func GetStatsSingle(viewId string) ([]float64, error) {
	rows, err := query("SELECT AVG(mood), AVG(happy), AVG(surprised), AVG(angry), AVG(disgusted), AVG(afraid), AVG(sad), AVG(engagement) FROM video_view_stats WHERE view_time_id IN (SELECT B.id FROM video_view AS A, video_view_time AS B WHERE A.view_id = ? AND A.view_id = B.view_id)", viewId)
	if err != nil {
		return nil, &InternalError{error: err}
	}
	defer rows.Close()

	emotionValues := make([]float64, 8, 8)
	emotionSqlValues := make([]sql.NullFloat64, 8, 8)

	if rows.Next() {
		err = rows.Scan(
			&emotionSqlValues[0],
			&emotionSqlValues[1],
			&emotionSqlValues[2],
			&emotionSqlValues[3],
			&emotionSqlValues[4],
			&emotionSqlValues[5],
			&emotionSqlValues[6],
			&emotionSqlValues[7],
		)

		if err != nil {
			return nil, &InternalError{error: err}
		}

		for idx, value := range emotionSqlValues {
			if value.Valid {
				emotionValues[idx] = value.Float64
			} else {
				if idx == 7 {
					emotionValues[idx] = 1
				} else {
					emotionValues[idx] = 0
				}
			}
		}

		return emotionValues, nil
	}

	return nil, errors.New("Mood and emotions query returned 0 rows")
}

func GetInstantStatsSingle(viewId string, startTime float64, endTime float64) ([]float64, error) {
	rows, err := query("SELECT AVG(mood), AVG(happy), AVG(surprised), AVG(angry), AVG(disgusted), AVG(afraid), AVG(sad), AVG(engagement) FROM video_view_stats WHERE view_time_id IN (SELECT B.id FROM video_view AS A, video_view_time AS B WHERE A.view_id = ? AND A.view_id = B.view_id AND B.time >= ? AND B.time < ?)", viewId, startTime, endTime)
	if err != nil {
		return nil, &InternalError{error: err}
	}
	defer rows.Close()

	emotionValues := make([]float64, 8, 8)
	emotionSqlValues := make([]sql.NullFloat64, 8, 8)

	if rows.Next() {
		err = rows.Scan(
			&emotionSqlValues[0],
			&emotionSqlValues[1],
			&emotionSqlValues[2],
			&emotionSqlValues[3],
			&emotionSqlValues[4],
			&emotionSqlValues[5],
			&emotionSqlValues[6],
			&emotionSqlValues[7],
		)

		if err != nil {
			return nil, &InternalError{error: err}
		}

		for idx, value := range emotionSqlValues {
			if value.Valid {
				emotionValues[idx] = value.Float64
			} else {
				if idx == 7 {
					emotionValues[idx] = 1
				} else {
					emotionValues[idx] = 0
				}
			}
		}

		return emotionValues, nil
	}

	return nil, errors.New("Mood and emotions instant query returned 0 rows")
}

func GetInstantViewedCountSingle(viewId string, startTime float64, endTime float64) (int64, error) {
	rows, err := query("SELECT COUNT(DISTINCT B.id) FROM video_view AS A, video_view_time AS B WHERE A.view_id = ? AND A.view_id = B.view_id AND B.time >= ? AND B.time < ?", viewId, startTime, endTime)
	if err != nil {
		return 0, &InternalError{error: err}
	}
	defer rows.Close()

	if rows.Next() {
		var count sql.NullInt64
		err = rows.Scan(&count)

		if err != nil {
			return 0, &InternalError{error: err}
		}

		if count.Valid {
			return count.Int64, nil
		} else {
			return 0, nil
		}
	}

	return 0, errors.New("COUNT(*) returned 0 rows")
}