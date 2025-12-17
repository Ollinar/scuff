



CREATE TABLE t_archive (
  c_id INTEGER PRIMARY KEY AUTOINCREMENT,
  c_path TEXT NOT NULL,
  c_size INTEGER NOT NULL,
  c_modtime INTEGER NOT NULL,
  c_type TEXT NOT NULL,
  c_partialhash TEXT NOT NULL
);

CREATE INDEX idx_t_archive_partialHash ON t_archive(c_partialhash);
CREATE INDEX idx_t_archive_path ON t_archive(c_path);

CREATE TABLE t_file (
  c_id INTEGER PRIMARY KEY AUTOINCREMENT,
  c_path TEXT NOT NULL,
  c_modtime INTEGER NOT NULL,
  c_mime TEXT NOT NULL,
  c_size INTEGER NOT NULL,
  c_archiveId INTEGER NOT NULL,
  FOREIGN KEY(c_archiveId) REFERENCES t_archive(c_id) ON DELETE CASCADE
);

CREATE TABLE t_chapter (
  c_id INTEGER PRIMARY KEY AUTOINCREMENT,
  c_name TEXT NOT NULL,
  c_description TEXT NOT NULL DEFAULT "",
  c_coverPageId INTEGER,
  FOREIGN KEY(c_coverPageId) REFERENCES t_page(c_id) ON DELETE CASCADE
);

CREATE TABLE t_page (
  c_id INTEGER PRIMARY KEY AUTOINCREMENT,
  c_width INTEGER NOT NULL,
  c_height INTEGER NOT NULL,
  c_isSpread BOOLEAN NOT NULL DEFAULT false,
  c_fileId INTEGER NOT NULL,
  c_pageName TEXT NOT NULL,
  FOREIGN KEY(c_fileId) REFERENCES t_file(c_id) ON DELETE CASCADE
);

CREATE TABLE t_tag (
  c_id INTEGER PRIMARY KEY AUTOINCREMENT,
  c_namespace TEXT NOT NULL,
  c_label TEXT NOT NULL,
  UNIQUE (c_namespace,c_label)
);

CREATE TABLE t_chapterTag (
  c_chapterId INTEGER,
  c_tagId INTEGER,
  PRIMARY KEY(c_chapterId,c_tagId),
  FOREIGN KEY(c_chapterId) REFERENCES t_chapter(c_id) ON DELETE CASCADE,
  FOREIGN KEY(c_tagId) REFERENCES t_tag(c_id) ON DELETE CASCADE
);

CREATE TABLE t_chapterPage (
 c_chapterId INTEGER NOT NULL,
 c_pageId INTEGER NOT NULL,
 c_pageNumber INTEGER NOT NULL,
 PRIMARY KEY(c_chapterId,c_pageNumber),
 FOREIGN KEY(c_chapterId) REFERENCES t_chapter(c_id) ON DELETE CASCADE,
 FOREIGN KEY(c_pageId) REFERENCES t_page(c_id) ON DELETE CASCADE
);

CREATE TABLE t_series (
  c_id INTEGER PRIMARY KEY AUTOINCREMENT,
  c_title TEXT NOT NULL,
  c_description TEXT NOT NULL
);


CREATE TABLE t_seriesChapter (
  c_chapterId INTEGER NOT NULL,
  c_seriesId INTEGER NOT NULL,
  c_chapterNumber INTEGER NOT NULL,
  PRIMARY KEY(c_seriesId,c_chapterNumber),
  FOREIGN KEY(c_chapterId) REFERENCES t_chapter(c_id) ON DELETE CASCADE,
  FOREIGN KEY(c_seriesId) REFERENCES t_series(c_id) ON DELETE CASCADE
);

CREATE TABLE t_pluginConfig (
  c_name TEXT,
  c_version TEXT,
  c_configJson TEXT NOT NULL,
  PRIMARY KEY(c_name,c_version)
);
