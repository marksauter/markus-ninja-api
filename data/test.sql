CREATE OR REPLACE VIEW event_master AS
SELECT
  event.*,
  CASE 
    WHEN course_event IS NOT NULL THEN
      json_build_object(
        'action', course_event.action,
        'course_id', course_event.course_id
      )
    WHEN lesson_comment_event IS NOT NULL THEN
      json_build_object(
        'action', lesson_comment_event.action,
        'comment_id', lesson_comment_event.comment_id,
        'lesson_id', lesson_comment_event.lesson_id
      )
    WHEN lesson_event_master IS NOT NULL THEN
      json_build_object(
        'action', lesson_event_master.action,
        'comment_id', lesson_event_master.comment_id,
        'course_id', lesson_event_master.course_id,
        'label_id', lesson_event_master.label_id,
        'lesson_id', lesson_event_master.lesson_id,
        'rename', lesson_event_master.rename,
        'source_id', lesson_event_master.source_id
      )
    WHEN study_event IS NOT NULL THEN
      json_build_object(
        'action', study_event.action,
        'study_id', study_event.study_id
      )
    WHEN user_asset_comment_event IS NOT NULL THEN
      json_build_object(
        'action', user_asset_comment_event.action,
        'comment_id', user_asset_comment_event.comment_id,
        'asset_id', user_asset_comment_event.asset_id
      )
    WHEN user_asset_event_master IS NOT NULL THEN
      json_build_object(
        'action', user_asset_event_master.action,
        'comment_id', user_asset_event_master.comment_id,
        'asset_id', user_asset_event_master.asset_id,
        'rename', user_asset_event_master.rename,
        'source_id', user_asset_event_master.source_id
      )
    ELSE NULL
  END AS payload
FROM event
LEFT JOIN course_event ON event.type = 'CourseEvent' AND course_event.event_id = event.id
LEFT JOIN lesson_comment_event ON event.type = 'LessonCommentEvent' AND lesson_comment_event.event_id = event.id
LEFT JOIN lesson_event_master ON event.type = 'LessonEvent' AND lesson_event_master.event_id = event.id
LEFT JOIN user_asset_comment_event ON event.type = 'UserAssetCommentEvent' AND user_asset_comment_event.event_id = event.id
LEFT JOIN user_asset_event_master ON event.type = 'UserAssetEvent' AND user_asset_event_master.event_id = event.id
LEFT JOIN study_event ON event.type = 'StudyEvent' AND study_event.event_id = event.id;

