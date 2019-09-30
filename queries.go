package main

const (
	httpQuery = ` FROM
(
SELECT 
"_http_requests"."action" AS "_http_requests_action",
"_http_requests"."http_referer" AS "_http_requests_http_referer",
"_http_requests"."http_user_agent" AS "_http_requests_http_user_agent",
"_http_requests"."http_request_uri" AS "_http_requests_http_request_uri",
"_http_requests"."created_at" AS "_http_requests_created_at",
"_http_requests"."completed_at" AS "_http_requests_completed_at",
"_http_requests"."worker" AS "_http_requests_worker",
"_http_requests"."vizql_session" AS "_http_requests_vizql_session",
"_http_requests"."user_ip" AS "_http_requests_user_ip",
"_http_requests"."currentsheet" AS "currentsheet",

"_Table2"."_users_id" AS "_users_id",
"_Table2"."name" AS "_users_name",
"_Table2"."login_at" AS "_users_login_at",
"_Table2"."friendly_name" AS "_users_friendly_name",
"_Table2"."system_user_id" AS "_users_system_user_id",
"_Table2"."domain_name" AS "_users_domain_name",
"_Table2"."domain_short_name" AS "_users_domain_short_name",

"_Table3"."_views_id" AS "_views_id",
"_Table3"."_views_name" AS "_views_name",
"_Table3"."view_url" AS "_views_view_url",
"_Table3"."_views_created_at" AS "_views_created_at",
"_Table3"."_views_title" AS "_views_title",
"_Table3"."_workbooks_id" AS "_workbooks_id",
"_Table3"."_workbooks_name" AS "_workbooks_name",
"_Table3"."workbook_url" AS "_workbooks_workbook_url",
"_Table3"."_workbooks_created_at" AS "_workbooks_created_at",
"_Table3"."updated_at" AS "_workbooks_updated_at",
"_Table3"."_workbooks_owner_id" AS "_workbooks_owner_id",
"_Table3"."size" AS "_workbooks_size",
"_Table3"."_workbooks_owner_name" AS "_workbooks_owner_name",
"_Table3"."project_name" AS "_workbooks_project_name",

"_Table4"."_http_requests_max_completed_at" AS "_http_requests_max_completed_at",
"_Table4"."_http_requests_min_created_at" AS "_http_requests_min_created_at"

FROM "public"."_http_requests" "_http_requests" 
LEFT JOIN 
(
SELECT 
  "_http_requests"."vizql_session" AS "vizql_session",
  "_users"."id" AS "_users_id",
  "_users"."name" AS "name",
  "_users"."login_at" AS "login_at",
  "_users"."friendly_name" AS "friendly_name",
  "_users"."licensing_role_id" AS "licensing_role_id",
  "_users"."licensing_role_name" AS "licensing_role_name",
  "_users"."domain_id" AS "domain_id",
  "_users"."system_user_id" AS "system_user_id",
  "_users"."domain_name" AS "domain_name",
  "_users"."domain_short_name" AS "domain_short_name",
  "_users"."site_id" AS "_users_site_id"
FROM "public"."_http_requests" "_http_requests" 
INNER JOIN "public"."_users" "_users" ON ("_http_requests"."user_id"="_users"."id")
GROUP BY  
"_http_requests"."vizql_session",
"_users"."id", 
"_users"."name",  
"_users"."login_at", 
"_users"."friendly_name",  
"_users"."licensing_role_id",  
"_users"."licensing_role_name", 
"_users"."domain_id", 
"_users"."system_user_id",
"_users"."domain_name", 
"_users"."domain_short_name", 
"_users"."site_id"
) 
as "_Table2"
ON ("_http_requests"."vizql_session" = "_Table2"."vizql_session")


LEFT JOIN
(
SELECT 
  "_views"."id" AS "_views_id",
  "_views"."name" AS "_views_name",
  "_views"."view_url" AS "view_url",
  "_views"."created_at" AS "_views_created_at",
  "_views"."owner_id" AS "owner_id",
  "_views"."owner_name" AS "owner_name",
  "_views"."workbook_id" AS "workbook_id",
  "_views"."index" AS "index",
  "_views"."title" AS "_views_title",
  "_views"."caption" AS "caption",
  "_views"."site_id" AS "_views_site_id",
  "_workbooks"."id" AS "_workbooks_id",
  "_workbooks"."name" AS "_workbooks_name",
  "_workbooks"."workbook_url" AS "workbook_url",
  "_workbooks"."created_at" AS "_workbooks_created_at",
  "_workbooks"."updated_at" AS "updated_at",
  "_workbooks"."owner_id" AS "_workbooks_owner_id",
  "_workbooks"."project_id" AS "project_id",
  "_workbooks"."size" AS "size",
  "_workbooks"."view_count" AS "view_count",
  "_workbooks"."owner_name" AS "_workbooks_owner_name",
  "_workbooks"."project_name" AS "project_name",
  "_workbooks"."system_user_id" AS "_workbooks_system_user_id",
  "_workbooks"."site_id" AS "_workbooks_site_id"
FROM "public"."_views" "_views"
INNER JOIN "public"."_workbooks" "_workbooks" ON ("_views"."workbook_id" = "_workbooks"."id")
) 
as "_Table3" 
ON ("_http_requests"."currentsheet" = "_Table3"."view_url")


LEFT JOIN 
(
SELECT 
"_http_requests"."vizql_session" AS "vizql_session",
max("_http_requests"."completed_at") AS "_http_requests_max_completed_at",
min("_http_requests"."created_at") AS "_http_requests_min_created_at",
max("_http_requests"."created_at")-min("_http_requests"."created_at")at time zone 'GMT' As "_http_requests_session time"

FROM "public"."_http_requests" "_http_requests" 
GROUP BY  
"_http_requests"."vizql_session"
) 
as "_Table4"
ON ("_http_requests"."vizql_session" = "_Table4"."vizql_session")
)
as dataset `
)
