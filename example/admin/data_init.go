package admin

import (
	"fmt"

	"github.com/qor5/admin/v3/example/models"
	"github.com/theplant/gofixtures"
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
		panic(fmt.Errorf("failed to obtain the admin role ID! %w", err))
	}

	// subQuery for finding the ids of these demo users
	subQuery := db.Table("users").Select("id").Where("account in (?)", reservedAccount)
	// obtain the user ids to be reserved
	var reservedUserIds []int
	err = db.Table("user_role_join").Group("user_id").
		Having("user_id IN (?) or COUNT(CASE WHEN role_id = (?) then 1 end)=1", subQuery, adminRoleID).
		Pluck("user_id", &reservedUserIds).Error
	if err != nil {
		panic(fmt.Errorf("failed to obtain the user ids to be retained! %w", err))
	}

	// First delete the data in the user_role_join table, then delete the data in the users table.
	// Due to foreign key constraints, it is not possible to delete data from the users table first.
	err = db.Exec("DELETE FROM user_role_join WHERE user_id NOT IN (?)", reservedUserIds).Error
	if err != nil {
		panic(fmt.Errorf("failed to delete public user related record in user_role_join table! %w", err))
	}
	err = db.Exec("DELETE FROM users WHERE id NOT IN (?)", reservedUserIds).Error
	if err != nil {
		panic(fmt.Errorf("failed to delete public user in users table! %w", err))
	}
}

// InitDB initializes the database with some initial data.
func InitDB(db *gorm.DB, tables []string) {
	var err error
	dbr, err := db.DB()
	if err != nil {
		panic(err)
	}

	// Page Builder
	PageBuilderExampleData.TruncatePut(dbr)
	// Orders
	OrdersExampleData.TruncatePut(dbr)
	// Workers
	WorkersExampleData.TruncatePut(dbr)
	// Categories
	CategoriesExampleData.TruncatePut(dbr)
	// InputDemos
	InputDemosExampleData.TruncatePut(dbr)
	// Posts
	PostsExampleData.TruncatePut(dbr)
	// NestedFieldDemos
	NestedFieldExampleData.TruncatePut(dbr)
	// ListModels
	ListModelsExampleData.TruncatePut(dbr)
	// MicrositeModels
	MicroSitesExampleData.TruncatePut(dbr)
	// Products
	ProductsExampleData.TruncatePut(dbr)
	// Media Library
	MediaLibrariesExampleData.TruncatePut(dbr)
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

var PageBuilderExampleData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.container_contact_form (id, add_top_space, add_bottom_space, anchor_id, heading, text, send_button_text, form_button_text, message_placeholder, name_placeholder, email_placeholder, thankyou_message, action_url, privacy_policy) VALUES (1, true, true, '', 'Get in touch', 'Whatever the challenge, we want to help you solve it.', 'Send', 'Write us', 'Your message', 'Name', 'Email', 'Thank you for getting in touch, we will get back to you soon.', '', 'I have read and agree to the <a href="/privacy-policy/" target="_blank">privacy policy</a>.');


INSERT INTO public.container_footers (id, english_url, japanese_url) VALUES (1, '', '');


INSERT INTO public.container_headers (id, color) VALUES (1, '');
INSERT INTO public.container_headers (id, color) VALUES (2, 'black');


INSERT INTO public.container_headings (id, add_top_space, add_bottom_space, anchor_id, heading, font_color, background_color, link, link_text, link_display_option, text) VALUES (1, false, false, '', 'Trusted by top brands', 'blue', 'white', '/projects/', 'LEARN MORE ABOUT OUR PROJECTS', 'all', '<p>We make your goals, our goals. Our innovative systems have a proven track record, delivering standout results for top brands. And with our expert team, we form lasting partnerships for the long-term. I don''t like this edit page</p>');
INSERT INTO public.container_headings (id, add_top_space, add_bottom_space, anchor_id, heading, font_color, background_color, link, link_text, link_display_option, text) VALUES (2, true, false, '', 'What we do', 'blue', 'grey', '/what-we-do/', 'LEARN MORE', 'all', '<p><strong>From end-to-end solutions to consulting, we draw on decades of expertise to solve new challenges in e-commerce, content management, and digital innovation.</strong></p>');
INSERT INTO public.container_headings (id, add_top_space, add_bottom_space, anchor_id, heading, font_color, background_color, link, link_text, link_display_option, text) VALUES (3, false, false, '', 'Why clients choose us', 'blue', 'white', '/why-clients-choose-us/', 'LEARN MORE', 'desktop', '');


INSERT INTO public.container_images (id, add_top_space, add_bottom_space, anchor_id, image, background_color, transition_background_color) VALUES (1, false, false, '', '{"ID":4,"Url":"https://cdn.qor5.theplant-dev.com/system/media_libraries/4/file.jpg","VideoLink":"","FileName":"file.20211006224452.jpg","Description":"","FileSizes":{"@qor_preview":19981,"default":257343,"original":257343},"Width":2880,"Height":720}', '', '');
INSERT INTO public.container_images (id, add_top_space, add_bottom_space, anchor_id, image, background_color, transition_background_color) VALUES (2, false, true, '', '{"ID":9,"Url":"https://cdn.qor5.theplant-dev.com/system/media_libraries/9/file.png","VideoLink":"","FileName":"file.png","Description":"","FileSizes":{"@qor_preview":41622,"default":227103,"original":227103},"Width":1252,"Height":658}', 'grey', 'white');


INSERT INTO public.container_video_banners (id, add_top_space, add_bottom_space, anchor_id, video, background_video, mobile_background_video, video_cover, mobile_video_cover, heading, popup_text, text, link_text, link) VALUES (1, false, false, '', '{"ID":14,"Url":"https://cdn.qor5.theplant-dev.com/system/media_libraries/14/file.mov","VideoLink":"","FileName":"pc-mv.mov","Description":""}', '{"ID":14,"Url":"https://cdn.qor5.theplant-dev.com/system/media_libraries/14/file.mov","VideoLink":"","FileName":"pc-mv.mov","Description":""}', '{"ID":11,"Url":"https://cdn.qor5.theplant-dev.com/system/media_libraries/11/file.mov","VideoLink":"","FileName":"mob-mv.mov","Description":""}', '{"ID":15,"Url":"https://cdn.qor5.theplant-dev.com/system/media_libraries/15/file.jpg","VideoLink":"","FileName":"pc.jpg","Description":"","FileSizes":{"@qor_preview":33019,"default":646542,"original":646542},"Width":2560,"Height":1440}', '{"ID":12,"Url":"https://cdn.qor5.theplant-dev.com/system/media_libraries/12/file.jpg","VideoLink":"","FileName":"mob.jpg","Description":"","FileSizes":{"@qor_preview":33140,"default":465542,"original":465542},"Width":1536,"Height":2876}', e'Enterprise systems.
Startup speed.', '', 'Discover made-to-measure enterprise solutions combining the agility of a startup with seamless performance at scale. The perfect fit, delivered fast.', 'get in touch', '/contact/');


INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (1, '2023-03-03 06:21:07.782515 +00:00', '2023-03-03 06:21:07.782515 +00:00', null, 'Product', '/product', '', 'International');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (2, '2023-03-03 06:21:15.410972 +00:00', '2023-03-03 06:21:15.410972 +00:00', null, 'Order', '/order', '', 'International');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (3, '2023-03-03 06:21:31.605906 +00:00', '2023-03-03 06:21:31.605906 +00:00', null, 'Food', '/product/food', '', 'International');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (1, '2023-03-03 06:21:07.782515 +00:00', '2023-03-03 06:21:07.782515 +00:00', null, 'Product', '/product', '', 'China');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (1, '2023-03-03 06:21:07.782515 +00:00', '2023-03-03 06:21:07.782515 +00:00', null, 'Product', '/product', '', 'Japan');


INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (1, '2023-03-03 06:20:48.334178 +00:00', '2023-03-03 06:20:48.334178 +00:00', null, 1, 'tpl', 'Image', 1, 1, false, false, 'Image', 'International', 0);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (2, '2023-03-03 06:21:40.233601 +00:00', '2023-03-03 06:21:40.233601 +00:00', null, 1, '2023-03-03-v01', 'Header', 1, 1, false, false, 'Header', 'International', 0);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (3, '2023-03-03 06:21:42.275791 +00:00', '2023-03-03 06:21:42.275791 +00:00', '2023-03-03 06:21:54.868151 +00:00', 1, '2023-03-03-v01', 'Header', 2, 2, false, false, 'Header', 'International', 0);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (4, '2023-03-03 06:21:58.674323 +00:00', '2023-03-03 06:21:58.674323 +00:00', null, 1, '2023-03-03-v01', 'Video Banner', 1, 2, false, false, 'Video Banner', 'International', 0);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (5, '2023-03-03 06:22:46.641959 +00:00', '2023-03-03 06:22:46.641959 +00:00', null, 1, '2023-03-03-v01', 'Heading', 1, 3, false, false, 'Heading', 'International', 0);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (7, '2023-03-03 06:24:15.676928 +00:00', '2023-03-03 06:24:15.676928 +00:00', null, 1, '2023-03-03-v01', 'Heading', 2, 5, false, false, 'Heading', 'International', 0);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (9, '2023-03-03 06:25:41.972811 +00:00', '2023-03-03 06:25:41.972811 +00:00', null, 1, '2023-03-03-v01', 'Image', 2, 7, false, false, 'Image', 'International', 0);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (10, '2023-03-03 06:25:55.874078 +00:00', '2023-03-03 06:25:55.874078 +00:00', null, 1, '2023-03-03-v01', 'Heading', 3, 8, false, false, 'Heading', 'International', 0);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (13, '2023-03-03 06:27:54.022522 +00:00', '2023-03-03 06:28:27.625631 +00:00', null, 1, '2023-03-03-v01', 'ContactForm', 1, 11, true, false, 'ContactForm', 'International', 0);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (14, '2023-03-03 06:28:30.305332 +00:00', '2023-03-03 06:28:30.305332 +00:00', null, 1, '2023-03-03-v01', 'Footer', 1, 12, false, false, 'Footer', 'International', 0);


INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, seo, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code) VALUES (1, '2024-05-23 12:45:15.163414 +00:00', '2024-05-23 12:45:15.163414 +00:00', null, 'The Plant Homepage', '/', 0, '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}', 'draft', '', null, null, null, null, '2023-03-03-v01', '2023-03-03-v01', '', 'International');

INSERT INTO public.page_builder_templates (id, created_at, updated_at, deleted_at, name, description, locale_code) VALUES (1, '2023-03-03 06:20:43.281780 +00:00', '2023-03-03 06:20:43.281780 +00:00', null, 'Demo', '', 'International');


`, []string{
	"container_contact_form",
	"container_footers",
	"container_headers",
	"container_headings",
	"container_images",
	"container_video_banners",
	"page_builder_categories",
	"page_builder_containers",
	"page_builder_pages",
	"page_builder_templates",
}))

var OrdersExampleData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.orders (id, created_at, updated_at, deleted_at, source, status, delivery_method, payment_method, confirmed_at, order_items) VALUES (4, '2022-10-13 10:41:47.425000 +00:00', null, null, 'APP', 'Pending', 'TableDelivery', 'PayPay', '2022-11-07 12:12:52.696000 +00:00', null);
INSERT INTO public.orders (id, created_at, updated_at, deleted_at, source, status, delivery_method, payment_method, confirmed_at, order_items) VALUES (6, '2022-10-17 10:26:51.856000 +00:00', null, null, 'WEB', 'Authorised', 'TableDelivery', 'PayPay', '2022-11-07 12:12:56.180000 +00:00', null);
INSERT INTO public.orders (id, created_at, updated_at, deleted_at, source, status, delivery_method, payment_method, confirmed_at, order_items) VALUES (5, '2022-10-13 10:42:11.414000 +00:00', null, null, 'APP', 'Cancelled', 'TableDelivery', 'PayPay', '2022-11-07 12:12:55.410000 +00:00', null);
INSERT INTO public.orders (id, created_at, updated_at, deleted_at, source, status, delivery_method, payment_method, confirmed_at, order_items) VALUES (8, '2022-11-07 12:19:59.612000 +00:00', null, null, 'APP', 'Sending', 'TableDelivery', 'CreditCard', '2022-11-07 12:20:20.468000 +00:00', null);
INSERT INTO public.orders (id, created_at, updated_at, deleted_at, source, status, delivery_method, payment_method, confirmed_at, order_items) VALUES (9, '2022-11-07 12:20:00.352000 +00:00', null, null, 'APP', 'CheckedIn', 'TableDelivery', 'CreditCard', '2022-11-07 12:20:21.212000 +00:00', null);
INSERT INTO public.orders (id, created_at, updated_at, deleted_at, source, status, delivery_method, payment_method, confirmed_at, order_items) VALUES (11, '2022-11-07 12:21:03.553000 +00:00', null, null, 'APP', 'Sending', 'TableDelivery', 'CreditCard', '2022-11-07 12:20:59.174000 +00:00', null);
INSERT INTO public.orders (id, created_at, updated_at, deleted_at, source, status, delivery_method, payment_method, confirmed_at, order_items) VALUES (7, '2022-11-07 12:19:57.671000 +00:00', null, null, 'APP', 'Sending', 'TableDelivery', 'CreditCard', '2022-11-07 12:20:19.556000 +00:00', null);
INSERT INTO public.orders (id, created_at, updated_at, deleted_at, source, status, delivery_method, payment_method, confirmed_at, order_items) VALUES (10, '2022-11-07 12:21:03.553000 +00:00', null, null, 'APP', 'Authorised', 'TableDelivery', 'CreditCard', '2022-11-07 12:20:59.174000 +00:00', null);

`, []string{"orders"}))

var WorkersExampleData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.qor_jobs (id, created_at, updated_at, deleted_at, job, status) VALUES (1, '2021-11-15 05:38:25.330081 +00:00', '2021-11-15 05:38:25.514704 +00:00', null, 'noArgJob', 'done');
INSERT INTO public.qor_jobs (id, created_at, updated_at, deleted_at, job, status) VALUES (34, '2022-10-08 03:15:48.245812 +00:00', '2022-10-14 07:16:05.216590 +00:00', null, 'scheduleJob', 'done');
INSERT INTO public.qor_jobs (id, created_at, updated_at, deleted_at, job, status) VALUES (2, '2021-12-07 13:31:07.383331 +00:00', '2021-12-07 13:31:12.457370 +00:00', null, 'progressTextJob', 'done');
INSERT INTO public.qor_jobs (id, created_at, updated_at, deleted_at, job, status) VALUES (3, '2022-01-10 11:51:44.495127 +00:00', '2022-01-10 11:51:44.622906 +00:00', null, 'scheduleJob', 'done');
INSERT INTO public.qor_jobs (id, created_at, updated_at, deleted_at, job, status) VALUES (67, '2022-10-20 02:38:34.139332 +00:00', '2022-10-20 02:38:39.247979 +00:00', null, 'errorJob', 'exception');
INSERT INTO public.qor_jobs (id, created_at, updated_at, deleted_at, job, status) VALUES (68, '2022-10-20 02:46:25.042928 +00:00', '2022-10-20 02:46:30.094506 +00:00', null, 'panicJob', 'exception');


INSERT INTO public.qor_job_instances (id, created_at, updated_at, deleted_at, qor_job_id, operator, job, status, args, context, progress, progress_text) VALUES (1, '2021-11-15 05:38:25.337004 +00:00', '2021-11-15 05:38:25.517271 +00:00', null, 1, null, 'noArgJob', 'done', 'null', null, 100, '');
INSERT INTO public.qor_job_instances (id, created_at, updated_at, deleted_at, qor_job_id, operator, job, status, args, context, progress, progress_text) VALUES (34, '2022-10-08 03:15:48.270563 +00:00', '2022-10-14 07:16:05.224650 +00:00', null, 34, '', 'scheduleJob', 'done', '{"F1":"f","ScheduleTime":"2022-10-14T07:16:00Z"}', '{"URL":"https://example.qor5.theplant-dev.com/admin/workers"}', 100, '');
INSERT INTO public.qor_job_instances (id, created_at, updated_at, deleted_at, qor_job_id, operator, job, status, args, context, progress, progress_text) VALUES (2, '2021-12-07 13:31:07.389003 +00:00', '2021-12-07 13:31:12.460350 +00:00', null, 2, null, 'progressTextJob', 'done', 'null', null, 100, '<a href="https://www.google.com">Download users</a>');
INSERT INTO public.qor_job_instances (id, created_at, updated_at, deleted_at, qor_job_id, operator, job, status, args, context, progress, progress_text) VALUES (3, '2022-01-10 11:51:44.506654 +00:00', '2022-01-10 11:51:44.631661 +00:00', null, 3, null, 'scheduleJob', 'done', '{"F1":"fda","ScheduleTime":null}', null, 100, '');
INSERT INTO public.qor_job_instances (id, created_at, updated_at, deleted_at, qor_job_id, operator, job, status, args, context, progress, progress_text) VALUES (67, '2022-10-20 02:38:34.152825 +00:00', '2022-10-20 02:38:39.251747 +00:00', null, 67, '', 'errorJob', 'exception', 'null', '{"URL":"https://example.qor5.theplant-dev.com/admin/workers"}', 0, 'imError');
INSERT INTO public.qor_job_instances (id, created_at, updated_at, deleted_at, qor_job_id, operator, job, status, args, context, progress, progress_text) VALUES (68, '2022-10-20 02:46:25.047450 +00:00', '2022-10-20 02:46:30.102953 +00:00', null, 68, '', 'panicJob', 'exception', 'null', '{"URL":"https://example.qor5.theplant-dev.com/admin/workers"}', 0, 'letsPanic');

`, []string{"qor_jobs", "qor_job_instances"}))

var CategoriesExampleData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.categories (id, created_at, updated_at, deleted_at, name, products, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version) VALUES (1, '2023-01-05 06:19:30.633000 +00:00', '2023-01-05 06:19:30.633000 +00:00', null, 'Demo', null, 'draft', '', null, null, null, null, '2023-01-05-v01', '', '');


`, []string{"categories"}))

var InputDemosExampleData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.input_demos (id, text_field1, text_area1, switch1, slider1, select1, range_slider1, radio1, file_input1, combobox1, checkbox1, autocomplete1, button_group1, chip_group1, item_group1, list_item_group1, slide_group1, color_picker1, date_picker1, date_picker_month1, time_picker1, media_library1, updated_at, created_at) VALUES (1, 'Demo', '', false, 0, '', null, '', '', '', false, '{}', '', '', '', '', '', '', '', '', '', null, '2023-01-05 06:21:36.488000 +00:00', '2023-01-05 06:21:36.488000 +00:00');


`, []string{"input_demos"}))

var PostsExampleData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.posts (id, created_at, updated_at, deleted_at, title, title_with_slug, seo, body, hero_image, body_image, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version) VALUES (1, '2023-01-05 06:23:55.553000 +00:00', '2023-01-05 06:23:55.553000 +00:00', null, 'Demo', 'demo', '{"Title":"","Description":"","Keywords":"","OpenGraphURL":"","OpenGraphType":"","OpenGraphImageURL":"","OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""},"OpenGraphMetadata":null,"EnabledCustomize":false}', '<p>Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Maecenas porttitor congue massa. Fusce posuere, magna sed pulvinar ultricies, purus lectus malesuada libero, sit amet commodo magna eros quis urna. Nunc viverra imperdiet enim. Fusce est. Vivamus a tellus. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Proin pharetra nonummy pede. Mauris et orci. Aenean nec lorem. In porttitor. Donec laoreet nonummy augue. Suspendisse dui purus, scelerisque at, vulputate vitae, pretium mattis, nunc. Mauris eget neque at sem venenatis eleifend. Ut nonummy.</p>', '{"ID":1,"Url":"//qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/1/file.jpeg","VideoLink":"","FileName":"demo image.jpeg","Description":"","FileSizes":{"@qor_preview":8917,"default":326350,"main":94913,"og":123973,"original":326350,"thumb":21199,"twitter-large":117784,"twitter-small":77615},"Width":750,"Height":1000}', null, 'draft', '', null, null, null, null, '2023-01-05-v01', '', '');



`, []string{"posts"}))

var NestedFieldExampleData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.customers (id, name) VALUES (1, 'Demo');


INSERT INTO public.addresses (id, customer_id, street, home_image, updated_at, created_at, status, online_url) VALUES (1, 1, 'Tokyo KDX Toranomon 1Chome Building 11F 1-10-5 Toranomon Minato-ku, Tokyo 〒105-0001', null, '2023-01-05 09:00:10.017949 +00:00', '2023-01-05 09:00:10.017949 +00:00', 'draft', '');
INSERT INTO public.addresses (id, customer_id, street, home_image, updated_at, created_at, status, online_url) VALUES (2, 1, 'Hangzhou Building #14 U3-2, No. 166  Lishui Rd, Gongshu Hangzhou, Zhejiang', null, '2023-01-05 09:00:10.017949 +00:00', '2023-01-05 09:00:10.017949 +00:00', 'draft', '');
INSERT INTO public.addresses (id, customer_id, street, home_image, updated_at, created_at, status, online_url) VALUES (3, 1, 'Canberra 73/30 Lonsdale Street, Braddon Canberra, ACT', null, '2023-01-05 09:00:10.017949 +00:00', '2023-01-05 09:00:10.017949 +00:00', 'draft', '');


INSERT INTO public.membership_cards (id, customer_id, number, valid_before) VALUES (1, 1, 0, null);




`, []string{"customers", "addresses", "membership_cards"}))

var ListModelsExampleData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.list_models (id, created_at, updated_at, deleted_at, title, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, page_number, position, list_deleted, list_updated, version, version_name, parent_version) VALUES (1, '2023-01-05 08:45:36.783000 +00:00', '2023-01-05 08:45:36.783000 +00:00', null, 'Demo', 'draft', '', null, null, null, null, 0, 0, false, false, '2023-01-05-v01', '', '');



`, []string{"list_models"}))

var MicroSitesExampleData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.micro_sites (id, created_at, updated_at, deleted_at, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, name, description, pre_path, package, files_list, unix_key) VALUES (1, '2023-01-05 08:49:45.695000 +00:00', '2023-01-05 08:49:45.695000 +00:00', null, 'draft', '', null, null, null, null, '2023-01-05-v01', '', '', 'Demo', '', '', '{"FileName":"","Url":""}', '', '');




`, []string{"micro_sites"}))

var ProductsExampleData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.products (id, created_at, updated_at, deleted_at, code, name, price, image, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version) VALUES (1, '2023-01-05 08:55:38.167000 +00:00', '2023-01-05 08:55:38.167000 +00:00', null, '001', 'cocacola', 5, '{"ID":34,"Url":"//qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/34/file.png","VideoLink":"","FileName":"3110-cocacola.png","Description":"","FileSizes":{"@qor_preview":35552,"default":18409,"original":18409,"thumb":11169},"Width":460,"Height":267}', 'draft', '', null, null, null, null, '2023-01-05-v01', '', '');


`, []string{"products"}))

var MediaLibrariesExampleData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file) VALUES (1, '2024-05-23 22:12:07.990058 +00:00', null, null, 'image', '{"FileName":"aigle.png","Url":"https://s3.ap-northeast-1.amazonaws.com/system/media_libraries/1/file.png","Width":320,"Height":84,"FileSizes":{"@qor_preview":17065,"default":3159,"original":3159},"Video":"","SelectedType":"","Description":""}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file) VALUES (2, '2024-05-23 22:12:07.990058 +00:00', null, null, 'image', '{"FileName":"asics.png","Url":"https://s3.ap-northeast-1.amazonaws.com/system/media_libraries/2/file.png","Width":254,"Height":84,"FileSizes":{"@qor_preview":15571,"default":3060,"original":3060},"Video":"","SelectedType":"","Description":""}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file) VALUES (3, '2024-05-23 22:12:07.990058 +00:00', null, null, 'image', '{"FileName":"file.20210903061739.png","Url":"https://s3.ap-northeast-1.amazonaws.com/system/media_libraries/3/file.png","Width":1722,"Height":196,"FileSizes":{"@qor_preview":627,"default":6887,"original":6887},"Video":"","SelectedType":"","Description":""}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file) VALUES (4, '2024-05-23 22:12:07.990058 +00:00', null, null, 'image', '{"FileName":"file.20211006224452.jpg","Url":"https://s3.ap-northeast-1.amazonaws.com/system/media_libraries/4/file.jpg","Width":2880,"Height":720,"FileSizes":{"@qor_preview":19981,"default":257343,"original":257343},"Video":"","SelectedType":"","Description":""}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file) VALUES (5, '2024-05-23 22:12:07.990058 +00:00', null, null, 'image', '{"FileName":"file.20211007041906.png","Url":"https://s3.ap-northeast-1.amazonaws.com/system/media_libraries/5/file.png","Width":481,"Height":741,"FileSizes":{"@qor_preview":79999,"default":234306,"original":234306},"Video":"","SelectedType":"","Description":""}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file) VALUES (6, '2024-05-23 22:12:07.990058 +00:00', null, null, 'image', '{"FileName":"file.20211007042027.png","Url":"https://s3.ap-northeast-1.amazonaws.com/system/media_libraries/6/file.png","Width":481,"Height":741,"FileSizes":{"@qor_preview":65623,"default":203098,"original":203098},"Video":"","SelectedType":"","Description":""}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file) VALUES (7, '2024-05-23 22:12:07.990058 +00:00', null, null, 'image', '{"FileName":"file.20211007042131.png","Url":"https://s3.ap-northeast-1.amazonaws.com/system/media_libraries/7/file.png","Width":481,"Height":741,"FileSizes":{"@qor_preview":64838,"default":189979,"original":189979},"Video":"","SelectedType":"","Description":""}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file) VALUES (8, '2024-05-23 22:12:07.990058 +00:00', null, null, 'image', '{"FileName":"file.20211007051449.png","Url":"https://s3.ap-northeast-1.amazonaws.com/system/media_libraries/8/file.png","Width":2880,"Height":1097,"FileSizes":{"@qor_preview":75734,"default":2236473,"original":2236473},"Video":"","SelectedType":"","Description":""}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file) VALUES (9, '2024-05-23 22:12:07.990058 +00:00', null, null, 'image', '{"FileName":"file.png","Url":"https://s3.ap-northeast-1.amazonaws.com/system/media_libraries/9/file.png","Width":1252,"Height":658,"FileSizes":{"@qor_preview":41622,"default":227103,"original":227103},"Video":"","SelectedType":"","Description":""}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file) VALUES (10, '2024-05-23 22:12:07.990058 +00:00', null, null, 'image', '{"FileName":"lacoste.png","Url":"https://s3.ap-northeast-1.amazonaws.com/system/media_libraries/10/file.png","Width":470,"Height":84,"FileSizes":{"@qor_preview":11839,"default":4714,"original":4714},"Video":"","SelectedType":"","Description":""}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file) VALUES (11, '2024-05-23 22:12:07.990058 +00:00', null, null, 'image', '{"FileName":"mob-mv.mov","Url":"https://s3.ap-northeast-1.amazonaws.com/system/media_libraries/11/file.mov","Video":"","SelectedType":"","Description":""}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file) VALUES (12, '2024-05-23 22:12:07.990058 +00:00', null, null, 'image', '{"FileName":"mob.jpg","Url":"https://s3.ap-northeast-1.amazonaws.com/system/media_libraries/12/file.jpg","Width":1536,"Height":2876,"FileSizes":{"@qor_preview":33140,"default":465542,"original":465542},"Video":"","SelectedType":"","Description":""}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file) VALUES (13, '2024-05-23 22:12:07.990058 +00:00', null, null, 'image', '{"FileName":"nhk.png","Url":"https://s3.ap-northeast-1.amazonaws.com/system/media_libraries/13/file.png","Width":202,"Height":84,"FileSizes":{"@qor_preview":14500,"default":2066,"original":2066},"Video":"","SelectedType":"","Description":""}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file) VALUES (14, '2024-05-23 22:12:07.990058 +00:00', null, null, 'image', '{"FileName":"pc-mv.mov","Url":"https://s3.ap-northeast-1.amazonaws.com/system/media_libraries/14/file.mov","Video":"","SelectedType":"","Description":""}');
INSERT INTO public.media_libraries (id, created_at, updated_at, deleted_at, selected_type, file) VALUES (15, '2024-05-23 22:12:07.990058 +00:00', null, null, 'image', '{"FileName":"pc.jpg","Url":"https://s3.ap-northeast-1.amazonaws.com/system/media_libraries/15/file.jpg","Width":2560,"Height":1440,"FileSizes":{"@qor_preview":33019,"default":646542,"original":646542},"Video":"","SelectedType":"","Description":""}');



`, []string{"media_libraries"}))
