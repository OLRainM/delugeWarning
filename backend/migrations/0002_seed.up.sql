-- 0002_seed: 基础种子数据（默认网格、预警模板、规则）

INSERT INTO grids (id, name, village, manager_user_id)
VALUES (1, '示范村一组', '示范村', 0)
ON CONFLICT (id) DO NOTHING;
SELECT setval(pg_get_serial_sequence('grids', 'id'), GREATEST((SELECT MAX(id) FROM grids), 1));

-- 预警模板：变量占位 {village} {device} {value} {level}
INSERT INTO templates (id, name, disaster_type, content_tpl, enabled) VALUES
(1, '洪水预警模板', 'flood',
 '【{level}预警】{village} {device} 监测水位已达 {value} 米，请注意防范，按指引转移避险。', true)
ON CONFLICT (id) DO NOTHING;
SELECT setval(pg_get_serial_sequence('templates', 'id'), GREATEST((SELECT MAX(id) FROM templates), 1));

-- 水位阈值规则：蓝2.5 / 黄3.0 / 橙3.5 / 红4.0
-- 低级别(蓝/黄)需人工复核，高级别(橙/红)直发
INSERT INTO rules (device_type, metric, operator, threshold, level, cooldown_sec, template_id, review_required, enabled) VALUES
('water_level', 'water_level', '>=', 2.5, 'blue',   600, 1, true,  true),
('water_level', 'water_level', '>=', 3.0, 'yellow', 600, 1, true,  true),
('water_level', 'water_level', '>=', 3.5, 'orange', 300, 1, false, true),
('water_level', 'water_level', '>=', 4.0, 'red',    180, 1, false, true)
ON CONFLICT DO NOTHING;

-- 示例设备
INSERT INTO devices (id, type, name, grid_id, status) VALUES
('dev-water-0001', 'water_level', '示范村河道水位计1', 1, 'online')
ON CONFLICT (id) DO NOTHING;
