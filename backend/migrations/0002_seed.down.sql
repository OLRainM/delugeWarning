-- 0002_seed down
DELETE FROM devices WHERE id = 'dev-water-0001';
DELETE FROM rules WHERE template_id = 1;
DELETE FROM templates WHERE id = 1;
DELETE FROM grids WHERE id = 1;
