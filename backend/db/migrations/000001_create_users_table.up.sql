CREATE TABLE user (
    id           INT          NOT NULL AUTO_INCREMENT,
    firebase_uid VARCHAR(128) NOT NULL,
    email        VARCHAR(255) NOT NULL,
    name         VARCHAR(100) NOT NULL,
    created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at   DATETIME     NULL     DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_user_firebase_uid (firebase_uid),
    UNIQUE KEY uq_user_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
