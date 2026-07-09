-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
  ADD COLUMN avatar_data MEDIUMBLOB COMMENT '头像二进制数据(JPEG压缩,≤150KB)',
  ADD COLUMN avatar_content_type VARCHAR(50) DEFAULT '' COMMENT '头像MIME类型';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
  DROP COLUMN avatar_data,
  DROP COLUMN avatar_content_type;
-- +goose StatementEnd
