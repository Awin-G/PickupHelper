-- +goose Up
-- +goose StatementBegin
INSERT IGNORE INTO `courier_companies` (`company_name`, `company_code`) VALUES
  ('顺丰速运', 'SF'),
  ('中通快递', 'ZTO'),
  ('圆通速递', 'YTO'),
  ('韵达快递', 'YUNDA'),
  ('申通快递', 'STO'),
  ('百世快递', 'HTKY'),
  ('邮政EMS', 'EMS'),
  ('京东物流', 'JD');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM `courier_companies` WHERE `company_code` IN ('SF','ZTO','YTO','YUNDA','STO','HTKY','EMS','JD');
-- +goose StatementEnd
