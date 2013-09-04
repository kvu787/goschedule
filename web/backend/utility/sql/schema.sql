-- NOTE: if the order and types of the fields do not map properly
-- to the structs defined backend/database, the backend/database
-- package will explode. 
-- TODO (kvu787): maybe use an external ORM package to avoid explosions

DROP TABLE depts;
DROP TABLE classes;
DROP TABLE sects;
DROP TABLE meeting_times;

CREATE TABLE depts (
    title text,
    abbreviation text CONSTRAINT pk_depts_abbreviation PRIMARY KEY,
    link text
);

CREATE TABLE classes (
    dept_abbreviation text,
    abbreviation_code text CONSTRAINT pk_classes_abbreviation_code PRIMARY KEY,
    abbreviation text,
    code text,
    title text,
    description text,
    index int
);

CREATE TABLE sects (
    class_dept_abbreviation text,
    restriction text,
    sln text CONSTRAINT pk_sects_sln PRIMARY KEY,
    section text,
    credit text,
    meeting_times text,
    instructor text,
    status text,
    taken_spots int,
    total_spots int,
    grades text,
    fee text,
    other text,
    info text
);

-- TODO (kvu787): make a proper 'has-and-belongs-to-many' relationship between sects and meeting_times
CREATE TABLE meeting_times (
    days text,
    time text,
    buliding text,
    room text
);