-- This SQL file will create the tables needed for the test case.

DROP TABLE IF EXISTS test_triangulation_info;

CREATE TABLE test_triangulation_info (
    id INTEGER PRIMARY KEY AUTOINCREMENT
    , name TEXT
);

DROP TABLE IF EXISTS test_triangulation_input_point;
CREATE TABLE test_triangulation_input_point (
    test_id INTEGER
    , "order" INTEGER
    , geometry POINT
    , FOREIGN KEY(test_id) REFERENCES test_triangulation_info(id)
);

DROP TABLE IF EXISTS "test_triangulation_expected_point";
CREATE TABLE "test_triangulation_expected_point" (
    test_id INTEGER
    , is_bad BOOLEAN DEFAULT false -- for good or bad points
    , "order" INTEGER
    , geometry POINT
    , FOREIGN KEY(test_id) REFERENCES test_triangulation_info(id)
);

DROP TABLE IF EXISTS "test_triangulation_expected_linestring";
CREATE TABLE "test_triangulation_expected_linestring" (
    test_id INTEGER
    , "order" INTEGER
    , geometry LINESTRING
    , FOREIGN KEY(test_id) REFERENCES test_triangulation_info(id)
);

DROP TABLE IF EXISTS "test_triangulation_expected_polygon";
CREATE TABLE "test_triangulation_expected_polygon" (
    test_id INTEGER
    , is_frame BOOLEAN DEFAULT false -- is part of the frame
    , type TEXT -- should be triangle, triangle:main, extent
    , "order" INTEGER
    , geometry LINESTRING
    , FOREIGN KEY(test_id) REFERENCES test_triangulation_info(id)
);
