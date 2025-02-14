-- +goose Up
CREATE TABLE Users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    picture TEXT DEFAULT 'default.jpg',
    role TEXT DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE Sessions (
    session_id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES Users(id)
);

CREATE TABLE Post (
    id INTEGER PRIMARY KEY,
    user_id UUID NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    category TEXT DEFAULT 'general',
    image TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES Users(id)
);

CREATE TABLE LikeDislikePost (
    id INTEGER PRIMARY KEY,
    user_id UUID NOT NULL,
    post_id INTEGER NOT NULL,
    like INTEGER DEFAULT 0,
    dislike INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES Users(id),
    FOREIGN KEY (post_id) REFERENCES Post(id),
    UNIQUE (user_id, post_id),
    CHECK (like + dislike <= 1)
);

CREATE TABLE Comment (
    id INTEGER PRIMARY KEY,
    user_id UUID NOT NULL,
    post_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES Users(id),
    FOREIGN KEY (post_id) REFERENCES Post(id)
);

CREATE TABLE LikeDislikeComment (
    id INTEGER PRIMARY KEY,
    user_id UUID NOT NULL,
    comment_id INTEGER NOT NULL,
    like INTEGER DEFAULT 0,
    dislike INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES Users(id),
    FOREIGN KEY (comment_id) REFERENCES Comment(id),
    UNIQUE (user_id, comment_id),            
    CHECK (like + dislike <= 1)
);

CREATE TABLE Report (
    id INTEGER PRIMARY KEY,
    user_id UUID NOT NULL,
    post_id INTEGER,
    comment_id INTEGER,
    reason TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES Users(id),
    FOREIGN KEY (post_id) REFERENCES Post(id),
    FOREIGN KEY (comment_id) REFERENCES Comment(id),
    CHECK ((post_id IS NOT NULL AND comment_id IS NULL) 
           OR (post_id IS NULL AND comment_id IS NOT NULL))
);

CREATE TABLE IF NOT EXISTS Categories (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,        
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS Catpostrel (
    cat_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    PRIMARY KEY (cat_id, post_id),           -- Clé primaire composite pour éviter les doublons
    FOREIGN KEY (cat_id) REFERENCES Categories(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (post_id) REFERENCES Post(id) ON DELETE CASCADE ON UPDATE CASCADE
);


CREATE TABLE IF NOT EXISTS Notification (
    id INTEGER PRIMARY KEY,
    user_id UUID NOT NULL,
    user_id2 UUID NOT NULL,
    post_id INTEGER,
    comment_id INTEGER,
    type TEXT NOT NULL,
    read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES Users(id),
    FOREIGN KEY (post_id) REFERENCES Post(id),
    FOREIGN KEY (comment_id) REFERENCES Comment(id),
    CHECK ((post_id IS NOT NULL AND comment_id IS NULL) 
           OR (post_id IS NULL AND comment_id IS NOT NULL))
);

CREATE TABLE IF NOT EXISTS Activity (
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- ID unique pour chaque activité
    user_id UUID NOT NULL,                -- L'utilisateur qui a réalisé l'activité
    activity_type TEXT NOT NULL,          -- Le type d'activité (e.g., 'post', 'like', 'comment', 'dislike')
    post_id INTEGER NULL,                 -- Le post associé à l'activité (si applicable)
    comment_id INTEGER NULL,              -- Le commentaire associé à l'activité (si applicable)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Date/heure de l'activité
    FOREIGN KEY (user_id) REFERENCES Users(id),
    FOREIGN KEY (post_id) REFERENCES Post(id),
    FOREIGN KEY (comment_id) REFERENCES Comment(id)
);

-- +goose Down
DROP TABLE IF EXISTS Report;
DROP TABLE IF EXISTS LikeDislikeComment;
DROP TABLE IF EXISTS Comment;
DROP TABLE IF EXISTS LikeDislikePost;
DROP TABLE IF EXISTS Post;
DROP TABLE IF EXISTS Users;
DROP TABLE IF EXISTS Categories;
DROP TABLE IF EXISTS Catpostrel;
DROP TABLE IF EXISTS Notification;
DROP TABLE IF EXISTS Activity;