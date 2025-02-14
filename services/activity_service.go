package services

import (
	"database/sql"
	"fmt"
	"forum/models"
	"time"

	"github.com/google/uuid"
)

type Activity struct {
	DB *sql.DB
}

func (a *Activity) CreateActivity(userID string, activityType string, postID int, commentID int) error {

	stmt := `
		INSERT INTO 
			Activity (user_id, activity_type, post_id, comment_id)
		VALUES 
			(?, ?, ?, ?);
	`
	_, err := a.DB.Exec(stmt, userID, activityType, postID, commentID)
	if err != nil {
		return err
	}
	return nil
}

func (a *Activity) DeleteActivity(userID string, activityType string, postID int, commentID int) error {

	stmt := `
		DELETE FROM 
			Activity 
		WHERE 
			user_id = ? AND activity_type = ? AND post_id = ? AND comment_id = ?;
	`
	_, err := a.DB.Exec(stmt, userID, activityType, postID, commentID)
	if err != nil {
		return err
	}
	return nil
}

func (a *Activity) LikeDislikeActivity(userID string, activityType string, postID int, commentID int) error {
	var existingActivityType string
	query := `
		SELECT activity_type 
		FROM Activity
		WHERE user_id = ? AND post_id = ? AND comment_id = ?;
	`
	err := a.DB.QueryRow(query, userID, postID, commentID).Scan(&existingActivityType)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		// Aucune interaction existante, insérer un nouveau like/dislike
		stmt := `
			INSERT INTO 
				Activity (user_id, activity_type, post_id, comment_id)
			VALUES 
				(?, ?, ?, ?);
		`
		_, err := a.DB.Exec(stmt, userID, activityType, postID, commentID)
		if err != nil {
			return err
		}
	} else if existingActivityType == activityType {
		// L'utilisateur a cliqué sur la même interaction (like ou dislike), donc on la supprime
		stmt := `
			DELETE FROM Activity
			WHERE user_id = ? AND post_id = ? AND comment_id = ?;
		`
		_, err := a.DB.Exec(stmt, userID, postID, commentID)
		if err != nil {
			return err
		}
	} else {
		// L'utilisateur change d'interaction (de like à dislike ou vice versa), mettre à jour
		stmt := `
			UPDATE Activity
			SET activity_type = ?
			WHERE user_id = ? AND post_id = ? AND comment_id = ?;
		`
		_, err := a.DB.Exec(stmt, activityType, userID, postID, commentID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Activity) DeleteActivityByPostID(postID int) error {
	stmt := `
		DELETE FROM Activity
		WHERE post_id = ?;
	`
	_, err := a.DB.Exec(stmt, postID)
	if err != nil {
		return err
	}
	return nil
}

func (a *Activity) GetPostbyUserId(userid string) error {
	// Démarre une transaction
	tx, err := a.DB.Begin()
	if err != nil {
		return fmt.Errorf("erreur lors de la création de la transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Relance la panique après le rollback
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	query := "SELECT id FROM Post WHERE user_id = ?"
	rows, err := tx.Query(query, userid)
	if err != nil {
		fmt.Printf("Erreur lors de la requête des posts pour l'utilisateur %s: %v\n", userid, err)
		return fmt.Errorf("erreur lors de la requête des posts pour l'utilisateur: %w", err)
	}
	defer rows.Close()

	var postIDs []int

	for rows.Next() {
		var postID int
		if err := rows.Scan(&postID); err != nil {
			fmt.Printf("Erreur lors du scan de l'ID du post pour l'utilisateur %s: %v\n", userid, err)
			return fmt.Errorf("erreur lors du scan de l'ID du post: %w", err)
		}
		postIDs = append(postIDs, postID)
	}

	if err := rows.Err(); err != nil {
		fmt.Printf("Erreur lors de l'itération des posts pour l'utilisateur %s: %v\n", userid, err)
		return fmt.Errorf("erreur lors de l'itération des posts: %w", err)
	}

	for _, postID := range postIDs {
		// Vérifie si l'activité existe déjà pour éviter les doublons
		checkQuery := "SELECT COUNT(1) FROM Activity WHERE user_id = ? AND activity_type = ? AND post_id = ? AND comment_id IS NULL"
		var count int
		err := tx.QueryRow(checkQuery, userid, "CreatedPost", postID).Scan(&count)
		if err != nil {
			fmt.Printf("Erreur lors de la vérification de l'existence de l'activité pour le post ID %d pour l'utilisateur %s: %v\n", postID, userid, err)
			return fmt.Errorf("erreur lors de la vérification de l'activité pour le post ID %d: %w", postID, err)
		}

		if count > 0 {
			// L'activité existe déjà, on passe au suivant
			continue
		}

		// Insère l'activité dans la base de données
		insertQuery := "INSERT INTO Activity (user_id, activity_type, post_id, comment_id) VALUES (?, ?, ?, ?)"
		_, err = tx.Exec(insertQuery, userid, "CreatedPost", postID, nil)
		if err != nil {
			fmt.Printf("Erreur lors de l'insertion de l'activité pour le post ID %d pour l'utilisateur %s: %v\n", postID, userid, err)
			return fmt.Errorf("erreur lors de l'insertion de l'activité pour le post ID %d: %w", postID, err)
		}
	}

	return err
}

func (a *Activity) GetAllActivityByUser(userid string) ([]models.ActivityPage, error) {
	fmt.Printf("UserID: %s\n", userid)

	stmt := `
        SELECT 
            Activity.id AS activity_id, 
            Activity.user_id AS activity_user_id,
            Activity.activity_type, 
            Activity.post_id AS activity_post_id, 
            Activity.comment_id AS activity_comment_id, 
            Activity.created_at AS activity_created_at,

            -- Activity User Info
            ActivityUser.id AS activity_user_userid,
            ActivityUser.username AS activity_user_username,
            ActivityUser.picture AS activity_user_picture,
            ActivityUser.role AS activity_user_role,
            ActivityUser.created_at AS activity_user_created_at,

            -- Post Info (for Activity's post)
            Post.id AS post_id,
            Post.title AS post_title,
            Post.content AS post_content,
            Post.image AS post_image,
            Post.user_id AS post_user_id,
            PostUser.username AS post_user_username,
            PostUser.picture AS post_user_picture,
            (SELECT COUNT(*) FROM LikeDislikePost WHERE LikeDislikePost.post_id = Post.id AND LikeDislikePost.like = 1) AS post_like_count,
            (SELECT COUNT(*) FROM LikeDislikePost WHERE LikeDislikePost.post_id = Post.id AND LikeDislikePost.dislike = 1) AS post_dislike_count,

            -- Comment Info
            Comment.id AS comment_id,
            Comment.content AS comment_content,
            Comment.created_at AS comment_created_at,
            Comment.user_id AS comment_user_id,
            CommentUser.username AS comment_user_username,
            CommentUser.picture AS comment_user_picture,
            (SELECT COUNT(*) FROM LikeDislikeComment WHERE LikeDislikeComment.comment_id = Comment.id AND LikeDislikeComment.like = 1) AS comment_like_count,
            (SELECT COUNT(*) FROM LikeDislikeComment WHERE LikeDislikeComment.comment_id = Comment.id AND LikeDislikeComment.dislike = 1) AS comment_dislike_count,

            -- Post Info for Comment's Post
            PostForComment.id AS comment_post_id,
            PostForComment.title AS comment_post_title,
            PostForComment.content AS comment_post_content,
            PostForComment.image AS comment_post_image,
            PostForComment.user_id AS comment_post_user_id,
            PostForCommentUser.username AS comment_post_user_username,
            PostForCommentUser.picture AS comment_post_user_picture,
            (SELECT COUNT(*) FROM LikeDislikePost WHERE LikeDislikePost.post_id = PostForComment.id AND LikeDislikePost.like = 1) AS comment_post_like_count,
            (SELECT COUNT(*) FROM LikeDislikePost WHERE LikeDislikePost.post_id = PostForComment.id AND LikeDislikePost.dislike = 1) AS comment_post_dislike_count

        FROM 
            Activity
        LEFT JOIN 
            Users AS ActivityUser ON Activity.user_id = ActivityUser.id
        LEFT JOIN 
            Post ON Activity.post_id = Post.id
        LEFT JOIN 
            Users AS PostUser ON Post.user_id = PostUser.id
        LEFT JOIN 
            Comment ON Activity.comment_id = Comment.id
        LEFT JOIN 
            Users AS CommentUser ON Comment.user_id = CommentUser.id
        LEFT JOIN
            Post AS PostForComment ON Comment.post_id = PostForComment.id
        LEFT JOIN
            Users AS PostForCommentUser ON PostForComment.user_id = PostForCommentUser.id
        WHERE 
            Activity.user_id = ?
        ORDER BY 
            Activity.created_at DESC;
    `

	rows, err := a.DB.Query(stmt, userid)
	if err != nil {
		return nil, fmt.Errorf("query error: %v", err)
	}
	defer rows.Close()

	var activities []models.ActivityPage
	likeModel := LikeModel{DB: a.DB}
	likeModelComment := LikeModelComment{DB: a.DB}

	for rows.Next() {
		var activity models.ActivityPage

		// Fields for UUIDs
		var activityUserIDStr string     // From Activity.user_id
		var activityUserUserIDStr string // From ActivityUser.id
		var postUserIDStr sql.NullString
		var commentUserIDStr sql.NullString
		var commentPostUserIDStr sql.NullString

		// Activity Post and Comment IDs
		var activityPostID sql.NullInt64    // From Activity.post_id
		var activityCommentID sql.NullInt64 // From Activity.comment_id

		// Fields for Activity User Info
		var activityUserUsername sql.NullString
		var activityUserPicture sql.NullString
		var activityUserRoles sql.NullString
		var activityUserCreatedAt sql.NullTime

		// Fields for Post Info (for Activity's post)
		var postID sql.NullInt64 // From Post.id
		var postTitle sql.NullString
		var postContent sql.NullString
		var postImage sql.NullString
		var postUserUsername sql.NullString
		var postUserPicture sql.NullString
		var postLikeCount sql.NullInt64
		var postDislikeCount sql.NullInt64

		// Fields for Comment Info
		var commentID sql.NullInt64 // From Comment.id
		var commentContent sql.NullString
		var commentCreatedAt sql.NullTime
		var commentUserUsername sql.NullString
		var commentUserPicture sql.NullString
		var commentLikeCount sql.NullInt64
		var commentDislikeCount sql.NullInt64

		// Fields for Comment's Post Info
		var commentPostID sql.NullInt64
		var commentPostTitle sql.NullString
		var commentPostContent sql.NullString
		var commentPostImage sql.NullString
		var commentPostUserUsername sql.NullString
		var commentPostUserPicture sql.NullString
		var commentPostLikeCount sql.NullInt64
		var commentPostDislikeCount sql.NullInt64

		// Scan the row
		err := rows.Scan(
			// Activity fields
			&activity.Id,
			&activityUserIDStr,
			&activity.ActivityType,
			&activityPostID,
			&activityCommentID,
			&activity.CreatedAt,

			// Activity User Info
			&activityUserUserIDStr,
			&activityUserUsername,
			&activityUserPicture,
			&activityUserRoles,
			&activityUserCreatedAt,

			// Post Info (for Activity's post)
			&postID,
			&postTitle,
			&postContent,
			&postImage,
			&postUserIDStr,
			&postUserUsername,
			&postUserPicture,
			&postLikeCount,
			&postDislikeCount,

			// Comment Info
			&commentID,
			&commentContent,
			&commentCreatedAt,
			&commentUserIDStr,
			&commentUserUsername,
			&commentUserPicture,
			&commentLikeCount,
			&commentDislikeCount,

			// Comment's Post Info
			&commentPostID,
			&commentPostTitle,
			&commentPostContent,
			&commentPostImage,
			&commentPostUserIDStr,
			&commentPostUserUsername,
			&commentPostUserPicture,
			&commentPostLikeCount,
			&commentPostDislikeCount,
		)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan error: %v", err)
		}

		// Parse UUIDs
		activityUserID, err := uuid.Parse(activityUserIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID in activityUserID: %v", err)
		}

		// Build Activity User
		activity.UserID = models.User{
			Id:       activityUserID,
			Username: activityUserUsername.String,
			Picture:  activityUserPicture.String,
			Roles:    activityUserRoles.String,
			CreatedAt: func() time.Time {
				if activityUserCreatedAt.Valid {
					return activityUserCreatedAt.Time
				}
				return time.Time{}
			}(),
		}

		// Build Post object if necessary
		if activityPostID.Valid && postID.Valid {
			postUserID, err := uuid.Parse(postUserIDStr.String)
			if err != nil {
				return nil, fmt.Errorf("invalid UUID in postUserID: %v", err)
			}

			// Get user action on the post
			userAction, err := likeModel.VerifyAction(fmt.Sprintf("%d", postID.Int64), userid)
			if err != nil {
				return nil, fmt.Errorf("error getting user action on post: %v", err)
			}

			activity.PostID = &models.Post{
				ID:      int(postID.Int64),
				Title:   postTitle.String,
				Content: postContent.String,
				Image: func() *string {
					if postImage.Valid {
						return &postImage.String
					}
					return nil
				}(),
				LikeCount:    int(postLikeCount.Int64),
				DislikeCount: int(postDislikeCount.Int64),
				UserAction:   userAction,
				UserID: models.User{
					Id:       postUserID,
					Username: postUserUsername.String,
					Picture:  postUserPicture.String,
				},
			}
		}

		// Build Comment object if necessary
		if activityCommentID.Valid && commentID.Valid {
			commentUserID, err := uuid.Parse(commentUserIDStr.String)
			if err != nil {
				return nil, fmt.Errorf("invalid UUID in commentUserID: %v", err)
			}
			// Get user action on the comment
			commentAction, err := likeModelComment.VerifyActionComment(fmt.Sprintf("%d", commentID.Int64), userid)
			if err != nil {
				return nil, fmt.Errorf("error getting user action on comment: %v", err)
			}

			// Parse commentPostUserID
			var commentPostUserID uuid.UUID
			if commentPostUserIDStr.Valid {
				commentPostUserID, err = uuid.Parse(commentPostUserIDStr.String)
				if err != nil {
					return nil, fmt.Errorf("invalid UUID in commentPostUserID: %v", err)
				}
			}

			// Build Post object for the comment's post
			commentPost := models.Post{
				ID:      int(commentPostID.Int64),
				Title:   commentPostTitle.String,
				Content: commentPostContent.String,
				Image: func() *string {
					if commentPostImage.Valid {
						return &commentPostImage.String
					}
					return nil
				}(),
				LikeCount:    int(commentPostLikeCount.Int64),
				DislikeCount: int(commentPostDislikeCount.Int64),
				UserID: models.User{
					Id:       commentPostUserID,
					Username: commentPostUserUsername.String,
					Picture:  commentPostUserPicture.String,
				},
			}

			activity.CommentID = &models.CommentActivity{
				ID:                  int(commentID.Int64),
				Content:             commentContent.String,
				LikeCountComment:    int(commentLikeCount.Int64),
				DislikeCountComment: int(commentDislikeCount.Int64),
				UserAction:          commentAction,
				UserID: models.User{
					Id:       commentUserID,
					Username: commentUserUsername.String,
					Picture:  commentUserPicture.String,
				},
				PostID:    commentPost,
				CreatedAt: commentCreatedAt.Time,
			}
		}

		activities = append(activities, activity)
	}

	fmt.Printf("Retrieved activities: %+v\n", activities)

	// Check for errors after row iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %v", err)
	}

	return activities, nil
}
