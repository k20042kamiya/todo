CREATE TABLE todo (
    id           INT          NOT NULL AUTO_INCREMENT,
    user_id      INT          NOT NULL,
    title        VARCHAR(255) NOT NULL,
    content      TEXT         NULL,
    due_date     DATE         NULL,
    is_completed BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at   DATETIME     NULL     DEFAULT NULL,
    PRIMARY KEY (id),
    KEY idx_todo_user_id (user_id),
    CONSTRAINT fk_todo_user_id FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
