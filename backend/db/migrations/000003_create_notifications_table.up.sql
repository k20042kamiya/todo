CREATE TABLE notification (
    id      INT                          NOT NULL AUTO_INCREMENT,
    todo_id INT                          NOT NULL,
    user_id INT                          NOT NULL,
    type    VARCHAR(32)                  NOT NULL,
    sent_at DATETIME                     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_notification_todo_id (todo_id),
    KEY idx_notification_user_id (user_id),
    CONSTRAINT fk_notification_todo_id FOREIGN KEY (todo_id) REFERENCES todo (id) ON DELETE CASCADE,
    CONSTRAINT fk_notification_user_id FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
