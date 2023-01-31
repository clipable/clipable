DROP INDEX IF EXISTS idx_user_trigram;
DROP FUNCTION IF EXISTS f_concat_ws(text, VARIADIC text[]);
DROP TABLE IF EXISTS "user";