package admin

import (
	"fmt"
	"time"

	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/media/media_library"
	"gorm.io/gorm"
)

func EmptyDB(db *gorm.DB, tables []string) {
	for _, name := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", name)).Error; err != nil {
			panic(err)
		}
	}
}

// ErasePublicUsersData erase all non-admin users but preserve the following three users
// qor@the-plant.com
// demo-editor@the-plant.com
// demo-viewer@the-plant.com
func ErasePublicUsersData(db *gorm.DB) {
	reservedAccount := []string{
		"qor@the-plant.com",
		"demo-editor@the-plant.com",
		"demo-viewer@the-plant.com",
	}

	var err error
	var adminRoleID int
	// obtain the admin role ID
	if err = db.Table("roles").Where("name = ?", models.RoleAdmin).
		Pluck("id", &adminRoleID).Error; err != nil {
		panic(fmt.Errorf("failed to obtain the admin role ID! %v", err))
	}

	// subQuery for finding the ids of these demo users
	subQuery := db.Table("users").Select("id").Where("account in (?)", reservedAccount)
	// obtain the user ids to be reserved
	var reservedUserIds []int
	err = db.Table("user_role_join").Group("user_id").
		Having("user_id IN (?) or COUNT(CASE WHEN role_id = (?) then 1 end)=1", subQuery, adminRoleID).
		Pluck("user_id", &reservedUserIds).Error
	if err != nil {
		panic(fmt.Sprintf("failed to obtain the user ids to be retained! %v", err))
	}

	// First delete the data in the user_role_join table, then delete the data in the users table.
	// Due to foreign key constraints, it is not possible to delete data from the users table first.
	err = db.Exec("DELETE FROM user_role_join WHERE user_id NOT IN (?)", reservedUserIds).Error
	if err != nil {
		panic(fmt.Errorf("failed to delete public user related record in user_role_join table! %v", err))
	}
	err = db.Exec("DELETE FROM users WHERE id NOT IN (?)", reservedUserIds).Error
	if err != nil {
		panic(fmt.Errorf("failed to delete public user in users table! %v", err))
	}
}

// InitDB initializes the database with some initial data.
func InitDB(db *gorm.DB, tables []string) {
	var err error

	// Page Builder
	if err = db.Exec(initPageBuilderSQL).Error; err != nil {
		panic(err)
	}
	// Orders
	if err = db.Exec(initOrdersSQL).Error; err != nil {
		panic(err)
	}
	// Workers
	if err = db.Exec(initWorkersSQL).Error; err != nil {
		panic(err)
	}
	// Categories
	if err = db.Exec(initCategoriesSQL).Error; err != nil {
		panic(err)
	}
	// InputDemos
	if err = db.Exec(initInputDemosSQL).Error; err != nil {
		panic(err)
	}
	// Posts
	if err = db.Exec(initPostsSQL).Error; err != nil {
		panic(err)
	}
	// NestedFieldDemos
	if err = db.Exec(initNestedFieldDemosSQL).Error; err != nil {
		panic(err)
	}
	// ListModels
	if err = db.Exec(initListModelsSQL).Error; err != nil {
		panic(err)
	}
	// MicrositeModels
	if err = db.Exec(initMicrositeModelsSQL).Error; err != nil {
		panic(err)
	}
	// Products
	if err = db.Exec(initProductsSQL).Error; err != nil {
		panic(err)
	}
	// Media Library
	now := time.Now()
	if err = db.Model(&media_library.MediaLibrary{}).Create(&[]map[string]interface{}{
		{"id": 1, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"aigle.png","Url":"%s","Width":320,"Height":84,"FileSizes":{"@qor_preview":17065,"default":3159,"original":3159},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/1/file.png")), "created_at": now},
		{"id": 2, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"asics.png","Url":"%s","Width":254,"Height":84,"FileSizes":{"@qor_preview":15571,"default":3060,"original":3060},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/2/file.png")), "created_at": now},
		{"id": 3, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"file.20210903061739.png","Url":"%s","Width":1722,"Height":196,"FileSizes":{"@qor_preview":627,"default":6887,"original":6887},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/3/file.png")), "created_at": now},
		{"id": 4, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"file.20211006224452.jpg","Url":"%s","Width":2880,"Height":720,"FileSizes":{"@qor_preview":19981,"default":257343,"original":257343},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/4/file.jpg")), "created_at": now},
		{"id": 5, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"file.20211007041906.png","Url":"%s","Width":481,"Height":741,"FileSizes":{"@qor_preview":79999,"default":234306,"original":234306},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/5/file.png")), "created_at": now},
		{"id": 6, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"file.20211007042027.png","Url":"%s","Width":481,"Height":741,"FileSizes":{"@qor_preview":65623,"default":203098,"original":203098},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/6/file.png")), "created_at": now},
		{"id": 7, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"file.20211007042131.png","Url":"%s","Width":481,"Height":741,"FileSizes":{"@qor_preview":64838,"default":189979,"original":189979},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/7/file.png")), "created_at": now},
		{"id": 8, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"file.20211007051449.png","Url":"%s","Width":2880,"Height":1097,"FileSizes":{"@qor_preview":75734,"default":2236473,"original":2236473},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/8/file.png")), "created_at": now},
		{"id": 9, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"file.png","Url":"%s","Width":1252,"Height":658,"FileSizes":{"@qor_preview":41622,"default":227103,"original":227103},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/9/file.png")), "created_at": now},
		{"id": 10, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"lacoste.png","Url":"%s","Width":470,"Height":84,"FileSizes":{"@qor_preview":11839,"default":4714,"original":4714},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/10/file.png")), "created_at": now},
		{"id": 11, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"mob-mv.mov","Url":"%s","Video":"","SelectedType":"","Description":""}`, composeS3Path("/11/file.mov")), "created_at": now},
		{"id": 12, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"mob.jpg","Url":"%s","Width":1536,"Height":2876,"FileSizes":{"@qor_preview":33140,"default":465542,"original":465542},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/12/file.jpg")), "created_at": now},
		{"id": 13, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"nhk.png","Url":"%s","Width":202,"Height":84,"FileSizes":{"@qor_preview":14500,"default":2066,"original":2066},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/13/file.png")), "created_at": now},
		{"id": 14, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"pc-mv.mov","Url":"%s","Video":"","SelectedType":"","Description":""}`, composeS3Path("/14/file.mov")), "created_at": now},
		{"id": 15, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"pc.jpg","Url":"%s","Width":2560,"Height":1440,"FileSizes":{"@qor_preview":33019,"default":646542,"original":646542},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/15/file.jpg")), "created_at": now},
	}).Error; err != nil {
		panic(err)
	}
	// Seq
	for _, name := range tables {
		if err := db.Exec(fmt.Sprintf("SELECT setval('%s_id_seq', (SELECT max(id) FROM %s));", name, name)).Error; err != nil {
			panic(err)
		}
	}
}

// composeS3Path to generate file path as https://cdn.qor5.com/system/media_libraries/236/file.jpeg.
func composeS3Path(filePath string) string {
	endPoint := s3Endpoint
	if endPoint == "" {
		endPoint = "https://cdn.qor5.com"
	}
	return fmt.Sprintf("%s/system/media_libraries%s", endPoint, filePath)
}

// GetNonIgnoredTableNames returns all table names except the ignored ones.
func GetNonIgnoredTableNames(db *gorm.DB) []string {
	ignoredTableNames := map[string]struct{}{
		"users":            {},
		"roles":            {},
		"user_role_join":   {},
		"login_sessions":   {},
		"qor_seo_settings": {},
	}

	var rawTableNames []string
	if err := db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema='public';").Scan(&rawTableNames).
		Error; err != nil {
		panic(err)
	}

	var tableNames []string
	for _, n := range rawTableNames {
		if _, ok := ignoredTableNames[n]; !ok {
			tableNames = append(tableNames, n)
		}
	}

	return tableNames
}

// Below is the SQL data for the above code.
const (
	initPageBuilderSQL = `


--
-- Data for Name: container_contact_form; Type: TABLE DATA; Schema: public; Owner: example
--

INSERT INTO public.container_contact_form VALUES (1, true, true, '', 'Get in touch', 'Whatever the challenge, we want to help you solve it.', 'Send', 'Write us', 'Your message', 'Name', 'Email', 'Thank you for getting in touch, we will get back to you soon.', '', 'I have read and agree to the <a href="/privacy-policy/" target="_blank">privacy policy</a>.');


--
-- Data for Name: container_footers; Type: TABLE DATA; Schema: public; Owner: example
--

INSERT INTO public.container_footers VALUES (1, '', '');


--
-- Data for Name: container_headers; Type: TABLE DATA; Schema: public; Owner: example
--

INSERT INTO public.container_headers VALUES (1, '');
INSERT INTO public.container_headers VALUES (2, 'black');


--
-- Data for Name: container_headings; Type: TABLE DATA; Schema: public; Owner: example
--

INSERT INTO public.container_headings VALUES (1, false, false, '', 'Trusted by top brands', 'blue', 'white', '/projects/', 'LEARN MORE ABOUT OUR PROJECTS', 'all', '<p>We make your goals, our goals. Our innovative systems have a proven track record, delivering standout results for top brands. And with our expert team, we form lasting partnerships for the long-term. I don''t like this edit page</p>');
INSERT INTO public.container_headings VALUES (2, true, false, '', 'What we do', 'blue', 'grey', '/what-we-do/', 'LEARN MORE', 'all', '<p><strong>From end-to-end solutions to consulting, we draw on decades of expertise to solve new challenges in e-commerce, content management, and digital innovation.</strong></p>');
INSERT INTO public.container_headings VALUES (3, false, false, '', 'Why clients choose us', 'blue', 'white', '/why-clients-choose-us/', 'LEARN MORE', 'desktop', '');


--
-- Data for Name: container_images; Type: TABLE DATA; Schema: public; Owner: example
--

INSERT INTO public.container_images VALUES (1, false, false, '', '{"ID":4,"Url":"https://cdn.qor5.theplant-dev.com/system/media_libraries/4/file.jpg","VideoLink":"","FileName":"file.20211006224452.jpg","Description":"","FileSizes":{"@qor_preview":19981,"default":257343,"original":257343},"Width":2880,"Height":720}', '', '');
INSERT INTO public.container_images VALUES (2, false, true, '', '{"ID":9,"Url":"https://cdn.qor5.theplant-dev.com/system/media_libraries/9/file.png","VideoLink":"","FileName":"file.png","Description":"","FileSizes":{"@qor_preview":41622,"default":227103,"original":227103},"Width":1252,"Height":658}', 'grey', 'white');



--
-- Data for Name: container_video_banners; Type: TABLE DATA; Schema: public; Owner: example
--

INSERT INTO public.container_video_banners VALUES (1, false, false, '', '{"ID":14,"Url":"https://cdn.qor5.theplant-dev.com/system/media_libraries/14/file.mov","VideoLink":"","FileName":"pc-mv.mov","Description":""}', '{"ID":14,"Url":"https://cdn.qor5.theplant-dev.com/system/media_libraries/14/file.mov","VideoLink":"","FileName":"pc-mv.mov","Description":""}', '{"ID":11,"Url":"https://cdn.qor5.theplant-dev.com/system/media_libraries/11/file.mov","VideoLink":"","FileName":"mob-mv.mov","Description":""}', '{"ID":15,"Url":"https://cdn.qor5.theplant-dev.com/system/media_libraries/15/file.jpg","VideoLink":"","FileName":"pc.jpg","Description":"","FileSizes":{"@qor_preview":33019,"default":646542,"original":646542},"Width":2560,"Height":1440}', '{"ID":12,"Url":"https://cdn.qor5.theplant-dev.com/system/media_libraries/12/file.jpg","VideoLink":"","FileName":"mob.jpg","Description":"","FileSizes":{"@qor_preview":33140,"default":465542,"original":465542},"Width":1536,"Height":2876}', 'Enterprise systems.
Startup speed.', '', 'Discover made-to-measure enterprise solutions combining the agility of a startup with seamless performance at scale. The perfect fit, delivered fast.', 'get in touch', '/contact/');


--
-- Data for Name: page_builder_categories; Type: TABLE DATA; Schema: public; Owner: example
--

INSERT INTO public.page_builder_categories VALUES (1, '2023-03-03 06:21:07.782515+00', '2023-03-03 06:21:07.782515+00', NULL, 'Product', '/product', '', 'International');
INSERT INTO public.page_builder_categories VALUES (2, '2023-03-03 06:21:15.410972+00', '2023-03-03 06:21:15.410972+00', NULL, 'Order', '/order', '', 'International');
INSERT INTO public.page_builder_categories VALUES (3, '2023-03-03 06:21:31.605906+00', '2023-03-03 06:21:31.605906+00', NULL, 'Food', '/product/food', '', 'International');
INSERT INTO public.page_builder_categories VALUES (1, '2023-03-03 06:21:07.782515+00', '2023-03-03 06:21:07.782515+00', NULL, 'Product', '/product', '', 'China');
INSERT INTO public.page_builder_categories VALUES (1, '2023-03-03 06:21:07.782515+00', '2023-03-03 06:21:07.782515+00', NULL, 'Product', '/product', '', 'Japan');

--
-- Data for Name: page_builder_containers; Type: TABLE DATA; Schema: public; Owner: example
--

INSERT INTO public.page_builder_containers VALUES (1, '2023-03-03 06:20:48.334178+00', '2023-03-03 06:20:48.334178+00', NULL, 1, 'tpl', 'Image', 1, 1, false, false, 'Image', 'International', 0);
INSERT INTO public.page_builder_containers VALUES (2, '2023-03-03 06:21:40.233601+00', '2023-03-03 06:21:40.233601+00', NULL, 1, '2023-03-03-v01', 'Header', 1, 1, false, false, 'Header', 'International', 0);
INSERT INTO public.page_builder_containers VALUES (3, '2023-03-03 06:21:42.275791+00', '2023-03-03 06:21:42.275791+00', '2023-03-03 06:21:54.868151+00', 1, '2023-03-03-v01', 'Header', 2, 2, false, false, 'Header', 'International', 0);
INSERT INTO public.page_builder_containers VALUES (4, '2023-03-03 06:21:58.674323+00', '2023-03-03 06:21:58.674323+00', NULL, 1, '2023-03-03-v01', 'Video Banner', 1, 2, false, false, 'Video Banner', 'International', 0);
INSERT INTO public.page_builder_containers VALUES (5, '2023-03-03 06:22:46.641959+00', '2023-03-03 06:22:46.641959+00', NULL, 1, '2023-03-03-v01', 'Heading', 1, 3, false, false, 'Heading', 'International', 0);
INSERT INTO public.page_builder_containers VALUES (7, '2023-03-03 06:24:15.676928+00', '2023-03-03 06:24:15.676928+00', NULL, 1, '2023-03-03-v01', 'Heading', 2, 5, false, false, 'Heading', 'International', 0);
INSERT INTO public.page_builder_containers VALUES (9, '2023-03-03 06:25:41.972811+00', '2023-03-03 06:25:41.972811+00', NULL, 1, '2023-03-03-v01', 'Image', 2, 7, false, false, 'Image', 'International', 0);
INSERT INTO public.page_builder_containers VALUES (10, '2023-03-03 06:25:55.874078+00', '2023-03-03 06:25:55.874078+00', NULL, 1, '2023-03-03-v01', 'Heading', 3, 8, false, false, 'Heading', 'International', 0);
INSERT INTO public.page_builder_containers VALUES (13, '2023-03-03 06:27:54.022522+00', '2023-03-03 06:28:27.625631+00', NULL, 1, '2023-03-03-v01', 'ContactForm', 1, 11, true, false, 'ContactForm', 'International', 0);
INSERT INTO public.page_builder_containers VALUES (14, '2023-03-03 06:28:30.305332+00', '2023-03-03 06:28:30.305332+00', NULL, 1, '2023-03-03-v01', 'Footer', 1, 12, false, false, 'Footer', 'International', 0);

--
-- Data for Name: page_builder_pages; Type: TABLE DATA; Schema: public; Owner: example
--

INSERT INTO public.page_builder_pages VALUES (1, '2023-03-03 06:20:35.886165+00', '2023-03-03 06:20:35.886165+00', NULL, 'The Plant Homepage', '/', 0, 'draft', '', NULL, NULL, NULL, NULL, '2023-03-03-v01', '', '', '', 'International');


--
-- Data for Name: page_builder_templates; Type: TABLE DATA; Schema: public; Owner: example
--

INSERT INTO public.page_builder_templates VALUES (1, '2023-03-03 06:20:43.28178+00', '2023-03-03 06:20:43.28178+00', NULL, 'Demo', '', 'International');
`

	initOrdersSQL = `
INSERT INTO public.orders VALUES (4, '2022-10-13 19:41:47.425+09', NULL, NULL, 'APP', 'Pending', 'TableDelivery', 'PayPay', '2022-11-07 21:12:52.696+09', NULL);
INSERT INTO public.orders VALUES (6, '2022-10-17 19:26:51.856+09', NULL, NULL, 'WEB', 'Authorised', 'TableDelivery', 'PayPay', '2022-11-07 21:12:56.18+09', NULL);
INSERT INTO public.orders VALUES (5, '2022-10-13 19:42:11.414+09', NULL, NULL, 'APP', 'Cancelled', 'TableDelivery', 'PayPay', '2022-11-07 21:12:55.41+09', NULL);
INSERT INTO public.orders VALUES (8, '2022-11-07 21:19:59.612+09', NULL, NULL, 'APP', 'Sending', 'TableDelivery', 'CreditCard', '2022-11-07 21:20:20.468+09', NULL);
INSERT INTO public.orders VALUES (9, '2022-11-07 21:20:00.352+09', NULL, NULL, 'APP', 'CheckedIn', 'TableDelivery', 'CreditCard', '2022-11-07 21:20:21.212+09', NULL);
INSERT INTO public.orders VALUES (11, '2022-11-07 21:21:03.553+09', NULL, NULL, 'APP', 'Sending', 'TableDelivery', 'CreditCard', '2022-11-07 21:20:59.174+09', NULL);
INSERT INTO public.orders VALUES (7, '2022-11-07 21:19:57.671+09', NULL, NULL, 'APP', 'Sending', 'TableDelivery', 'CreditCard', '2022-11-07 21:20:19.556+09', NULL);
INSERT INTO public.orders VALUES (10, '2022-11-07 21:21:03.553+09', NULL, NULL, 'APP', 'Authorised', 'TableDelivery', 'CreditCard', '2022-11-07 21:20:59.174+09', NULL);
`

	initWorkersSQL = `
INSERT INTO public.qor_jobs VALUES (1, '2021-11-15 14:38:25.330081+09', '2021-11-15 14:38:25.514704+09', NULL, 'noArgJob', 'done');
INSERT INTO public.qor_jobs VALUES (34, '2022-10-08 12:15:48.245812+09', '2022-10-14 16:16:05.21659+09', NULL, 'scheduleJob', 'done');
INSERT INTO public.qor_jobs VALUES (2, '2021-12-07 22:31:07.383331+09', '2021-12-07 22:31:12.45737+09', NULL, 'progressTextJob', 'done');
INSERT INTO public.qor_jobs VALUES (3, '2022-01-10 20:51:44.495127+09', '2022-01-10 20:51:44.622906+09', NULL, 'scheduleJob', 'done');
INSERT INTO public.qor_jobs VALUES (67, '2022-10-20 11:38:34.139332+09', '2022-10-20 11:38:39.247979+09', NULL, 'errorJob', 'exception');
INSERT INTO public.qor_jobs VALUES (68, '2022-10-20 11:46:25.042928+09', '2022-10-20 11:46:30.094506+09', NULL, 'panicJob', 'exception');

INSERT INTO public.qor_job_instances (id, created_at, updated_at, deleted_at, qor_job_id, job, status, args, progress, progress_text, operator, context) VALUES (1, '2021-11-15 05:38:25.337004 +00:00', '2021-11-15 05:38:25.517271 +00:00', NULL, 1, 'noArgJob', 'done', 'null', 100, '', NULL, NULL);
INSERT INTO public.qor_job_instances (id, created_at, updated_at, deleted_at, qor_job_id, job, status, args, progress, progress_text, operator, context) VALUES (34, '2022-10-08 03:15:48.270563 +00:00', '2022-10-14 07:16:05.224650 +00:00', NULL, 34, 'scheduleJob', 'done', '{"F1":"f","ScheduleTime":"2022-10-14T07:16:00Z"}', 100, '', '', '{"URL":"https://example.qor5.theplant-dev.com/admin/workers"}');
INSERT INTO public.qor_job_instances (id, created_at, updated_at, deleted_at, qor_job_id, job, status, args, progress, progress_text, operator, context) VALUES (2, '2021-12-07 13:31:07.389003 +00:00', '2021-12-07 13:31:12.460350 +00:00', NULL, 2, 'progressTextJob', 'done', 'null', 100, '<a href="https://www.google.com">Download users</a>', NULL, NULL);
INSERT INTO public.qor_job_instances (id, created_at, updated_at, deleted_at, qor_job_id, job, status, args, progress, progress_text, operator, context) VALUES (3, '2022-01-10 11:51:44.506654 +00:00', '2022-01-10 11:51:44.631661 +00:00', NULL, 3, 'scheduleJob', 'done', '{"F1":"fda","ScheduleTime":null}', 100, '', NULL, NULL);
INSERT INTO public.qor_job_instances (id, created_at, updated_at, deleted_at, qor_job_id, job, status, args, progress, progress_text, operator, context) VALUES (67, '2022-10-20 02:38:34.152825 +00:00', '2022-10-20 02:38:39.251747 +00:00', NULL, 67, 'errorJob', 'exception', 'null', 0, 'imError', '', '{"URL":"https://example.qor5.theplant-dev.com/admin/workers"}');
INSERT INTO public.qor_job_instances (id, created_at, updated_at, deleted_at, qor_job_id, job, status, args, progress, progress_text, operator, context) VALUES (68, '2022-10-20 02:46:25.047450 +00:00', '2022-10-20 02:46:30.102953 +00:00', NULL, 68, 'panicJob', 'exception', 'null', 0, 'letsPanic', '', '{"URL":"https://example.qor5.theplant-dev.com/admin/workers"}');
`

	initCategoriesSQL = `INSERT INTO "categories" ("created_at","updated_at","deleted_at","name","products","status","online_url","scheduled_start_at","scheduled_end_at","actual_start_at","actual_end_at","version_name","parent_version","version") VALUES ('2023-01-05 15:19:30.633','2023-01-05 15:19:30.633',NULL,'Demo',NULL,'draft','',NULL,NULL,NULL,NULL,'','','2023-01-05-v01') RETURNING "id","version"`

	initInputDemosSQL = `INSERT INTO "input_demos" ("text_field1","text_area1","switch1","slider1","select1","range_slider1","radio1","file_input1","combobox1","checkbox1","autocomplete1","button_group1","chip_group1","item_group1","list_item_group1","slide_group1","color_picker1","date_picker1","date_picker_month1","time_picker1","media_library1","updated_at","created_at") VALUES ('Demo','',FALSE,0,'',NULL,'','','',FALSE,'{""}','','','','','','','','','',NULL,'2023-01-05 15:21:36.488','2023-01-05 15:21:36.488') RETURNING "id"`

	initPostsSQL = `INSERT INTO "posts" ("created_at","updated_at","deleted_at","title","title_with_slug","seo","body","hero_image","body_image","status","online_url","scheduled_start_at","scheduled_end_at","actual_start_at","actual_end_at","version_name","parent_version","version") VALUES ('2023-01-05 15:23:55.553','2023-01-05 15:23:55.553',NULL,'Demo','demo','{"Title":"","Description":"","Keywords":"","OpenGraphURL":"","OpenGraphType":"","OpenGraphImageURL":"","OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""},"OpenGraphMetadata":null,"EnabledCustomize":false}','<p>Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Maecenas porttitor congue massa. Fusce posuere, magna sed pulvinar ultricies, purus lectus malesuada libero, sit amet commodo magna eros quis urna. Nunc viverra imperdiet enim. Fusce est. Vivamus a tellus. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Proin pharetra nonummy pede. Mauris et orci. Aenean nec lorem. In porttitor. Donec laoreet nonummy augue. Suspendisse dui purus, scelerisque at, vulputate vitae, pretium mattis, nunc. Mauris eget neque at sem venenatis eleifend. Ut nonummy.</p>','{"ID":1,"Url":"//qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/1/file.jpeg","VideoLink":"","FileName":"demo image.jpeg","Description":"","FileSizes":{"@qor_preview":8917,"default":326350,"main":94913,"og":123973,"original":326350,"thumb":21199,"twitter-large":117784,"twitter-small":77615},"Width":750,"Height":1000}',NULL,'draft','',NULL,NULL,NULL,NULL,'','','2023-01-05-v01') RETURNING "id","version"`

	initNestedFieldDemosSQL = `
INSERT INTO public.customers VALUES (1, 'Demo');

INSERT INTO public.addresses VALUES (1, 1, 'Tokyo KDX Toranomon 1Chome Building 11F 1-10-5 Toranomon Minato-ku, Tokyo 〒105-0001', NULL, '2023-01-05 09:00:10.017949+00', '2023-01-05 09:00:10.017949+00', 'draft', '');
INSERT INTO public.addresses VALUES (2, 1, 'Hangzhou Building #14 U3-2, No. 166  Lishui Rd, Gongshu Hangzhou, Zhejiang', NULL, '2023-01-05 09:00:10.017949+00', '2023-01-05 09:00:10.017949+00', 'draft', '');
INSERT INTO public.addresses VALUES (3, 1, 'Canberra 73/30 Lonsdale Street, Braddon Canberra, ACT', NULL, '2023-01-05 09:00:10.017949+00', '2023-01-05 09:00:10.017949+00', 'draft', '');

INSERT INTO public.membership_cards VALUES (1, 1, 0, NULL);
`

	initListModelsSQL = `INSERT INTO "list_models" ("created_at","updated_at","deleted_at","title","status","online_url","scheduled_start_at","scheduled_end_at","actual_start_at","actual_end_at","page_number","position","list_deleted","list_updated","version_name","parent_version","version") VALUES ('2023-01-05 17:45:36.783','2023-01-05 17:45:36.783',NULL,'Demo','draft','',NULL,NULL,NULL,NULL,0,0,FALSE,FALSE,'','','2023-01-05-v01') RETURNING "id","version"`

	initMicrositeModelsSQL = `INSERT INTO "micro_sites" ("name","description","created_at","updated_at","deleted_at",
"status","online_url","scheduled_start_at","scheduled_end_at","actual_start_at","actual_end_at","version_name","parent_version","pre_path","package","files_list","unix_key","version") VALUES ('Demo','','2023-01-05 17:49:45.695','2023-01-05 17:49:45.695',NULL,'draft','',NULL,NULL,NULL,NULL,'','','','{"FileName":"","Url":""}','','','2023-01-05-v01') RETURNING "id","version"`

	initProductsSQL = `INSERT INTO "products" ("created_at","updated_at","deleted_at","code","name","price","image","status","online_url","scheduled_start_at","scheduled_end_at","actual_start_at","actual_end_at","version_name","parent_version","version") VALUES ('2023-01-05 17:55:38.167','2023-01-05 17:55:38.167',NULL,'001','cocacola',5,'{"ID":34,"Url":"//qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/34/file.png","VideoLink":"","FileName":"3110-cocacola.png","Description":"","FileSizes":{"@qor_preview":35552,"default":18409,"original":18409,"thumb":11169},"Width":460,"Height":267}','draft','',NULL,NULL,NULL,NULL,'','','2023-01-05-v01') RETURNING "id","version"`
)
