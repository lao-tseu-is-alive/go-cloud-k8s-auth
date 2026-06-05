-- the next lines with create extension should be done by a superuser of db
-- I added those to the scripts/createLocalDBAndUser.sh
-- CREATE EXTENSION postgis;
-- CREATE EXTENSION unaccent;

CREATE SCHEMA template_4_your_project_name_db_schema;
CREATE TYPE template_4_your_project_name_status_type AS ENUM (
    'Planifié', -- This state refers to sometemplate_4_your_project_name that is in the early stages of development, where the design and construction plans have been finalized, but physical construction has not yet begun.
    'En Construction', -- Sometemplate_4_your_project_name is in the process of being built.
    'Utilisé', -- The template_4_your_project_name is now being utilized for its intended purpose.
    'Abandonné', -- The template_4_your_project_name is no longer in use, being utilized
    'Démoli'
    );
-- TO LIST THOSE ENUM VALUES:
--SELECT unnest(enum_range(NULL::template_4_your_project_name_status_type))
/*
-- Table structure for table template_4_your_project_name for project template-4-your-project-name
-- version : 0.0.4
*/
CREATE TABLE IF NOT EXISTS template_4_your_project_name_db_schema.template_4_your_project_name
(
    -- using Postgres Native UUID v4 128bit https://www.postgresql.org/docs/14/datatype-uuid.html
    -- this choice allows to do client side generation of the id UUID v4
    -- https://shekhargulati.com/2022/06/23/choosing-a-primary-key-type-in-postgres/
    id                 uuid    not null
        constraint pk_template_4_your_project_name primary key default gen_random_uuid(),
    type_id            integer not null,
    name               text    not null constraint template_4_your_project_name_db_schema_unique_name	unique
        constraint name_min_length check (length(btrim(name)) > 2),
    description        text,
    comment            text,
    external_id        integer,
    external_ref       text,
    build_at           timestamp,
    status             template_4_your_project_name_status_type,
    contained_by       uuid,
    contained_by_old   integer,  -- to simplify initial import of data
    inactivated        boolean   not null      default false,
    inactivated_time   timestamp,
    inactivated_by     integer,
    inactivated_reason text,
    validated          boolean          default false,
    validated_time     timestamp,
    validated_by       integer,
    managed_by         integer, --link to actor
    _created_at        timestamp        default now() not null,
    _created_by        integer not null,
    _last_modified_at  timestamp,
    _last_modified_by  integer,
    _deleted           boolean          default false,
    _deleted_at        timestamp,
    _deleted_by        integer,
    more_data          jsonb,
    text_search        tsvector
);

SELECT AddGeometryColumn('template_4_your_project_name_db_schema', 'template_4_your_project_name', 'position', 2056, 'POINT', 2);
CREATE INDEX idx_template_4_your_project_name_geom_gist ON template_4_your_project_name_db_schema.template_4_your_project_name USING gist (position);
CREATE INDEX idx_template_4_your_project_name_type_id ON template_4_your_project_name_db_schema.template_4_your_project_name (type_id);


--
-- Table structure for table `Typetemplate4YourProjectName` generated from model 'Typetemplate4YourProjectName'
--
CREATE TABLE IF NOT EXISTS template_4_your_project_name_db_schema.type_template_4_your_project_name
(
    id                 serial
        constraint pk_type_template_4_your_project_name primary key,
    name               text                    not null,
    description        text,
    comment            text,
    external_id        integer,
    table_name         text,
    geometry_type      text,
    inactivated        boolean   default false,
    inactivated_time   timestamp,
    inactivated_by     integer,
    inactivated_reason text,
    managed_by         integer,
    icon_path          text default '/img/gomarker_star_blue.png' not null,
    _created_at        timestamp default now() not null,
    _created_by        integer                 not null,
    _last_modified_at  timestamp,
    _last_modified_by  integer,
    _deleted           boolean   default false,
    _deleted_at        timestamp,
    _deleted_by        integer,
    more_data_schema   jsonb,
    text_search        tsvector
);

alter table template_4_your_project_name_db_schema.template_4_your_project_name
    add constraint template_4_your_project_name_type_template_4_your_project_name_id_fk
        foreign key (type_id) references template_4_your_project_name_db_schema.type_template_4_your_project_name;

-- #### BEGIN OF DATA FROM GOELAND
-- imported from Typetemplate4YourProjectName in goeland as of 2023-07-21
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (1, 'Rue', 'Artères,Rues & chemins', null, 1, 'ThiStreet', false, null, null, null, 6,
        '1998-05-07 13:39:11.000000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (2, 'Ville', 'Villes & communes', null, 2, 'ThiCity', false, null, null, null, 6, '1998-05-11 13:12:30.000000',
        6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (3, 'Parcelle', 'Parcelles', null, 3, 'ThiParcelle', false, null, null, null, 6, '1998-10-05 16:25:51.080000', 6,
        null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (4, 'Lieu-dit', 'Lieu-dit', null, 4, null, false, null, null, null, 6, '1999-02-16 09:32:20.340000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type, icon_path)
VALUES (5, 'Bâtiment', 'Bâtiments', null, 5, 'ThiBuilding', false, null, null, null, 6, '1999-02-23 14:54:27.060000', 6,
        null, null, false, null, null, null, 'bbox', '"/img/gomarker_building.png"');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (6, 'Mur de quai', 'Mur de quai (rives de lac)', null, 6, null, false, null, null, null, 6,
        '2000-03-17 08:10:05.143000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (7, 'Grue', 'Grue', null, 7, null, false, null, null, null, 6, '1999-12-16 15:52:12.943000', 6, null, null,
        false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (8, 'Pendoir', 'Pendoir à bateau', null, 8, null, false, null, null, null, 6, '1999-12-16 15:53:06.570000', 6,
        null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (9, 'Borne eau/electricité', 'Borne eau/electricité', null, 9, null, false, null, null, null, 6,
        '1999-12-16 15:58:00.367000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (10, 'Pompe', 'Pompe', null, 10, null, false, null, null, null, 6, '1999-12-16 15:59:09.180000', 6, null, null,
        false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (11, 'Amarrage', 'Amarrage', null, 11, null, false, null, null, null, 6, '2000-03-15 15:20:24.613000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (12, 'Bassin', 'Bassin ', null, 12, null, false, null, null, null, 6, '2000-03-15 15:21:57.097000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (13, 'Candélabre ', 'Candélabre ', null, 13, null, false, null, null, null, 6, '2000-03-15 15:22:53.660000', 6,
        null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (14, 'Caniveau ', 'Caniveau', null, 14, null, false, null, null, null, 6, '2000-03-15 15:24:37.457000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (15, 'Collecteur ', 'Collecteur ', null, 15, null, false, null, null, null, 6, '2000-03-15 15:26:28.613000', 6,
        null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (16, 'Echelle', 'Echelle', null, 16, null, false, null, null, null, 6, '2000-03-15 15:26:59.597000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (17, 'Enrochement', 'Enrochement ', null, 17, null, false, null, null, null, 6, '2000-03-15 15:28:35.517000', 6,
        null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (18, 'Estacade', 'Estacade', null, 18, null, false, null, null, null, 6, '2000-03-15 15:31:56.800000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (19, 'Ponton', 'Ponton ', null, 19, null, false, null, null, null, 6, '2000-03-15 15:32:45.673000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (20, 'Ouvrage spécial', 'Ouvrage spécial', null, 20, null, false, null, null, null, 6,
        '2000-03-17 08:10:47.720000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (21, 'Sculpture', 'Sculpture', null, 21, null, false, null, null, null, 6, '2000-03-17 10:11:51.550000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (22, 'Citerne', 'Citerne', null, 22, null, false, null, null, null, 6, '2000-03-17 10:12:29.737000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (23, 'Fontaine', 'Fontaine', null, 23, 'ThiFontaine', false, null, null, null, 6, '2000-03-17 10:13:33.707000',
        6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (24, 'Feux', 'Feux de signalisation', null, 24, null, false, null, null, null, 6, '2000-03-23 15:35:16.247000',
        6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (25, 'Glacis', 'Glacis', null, 25, null, false, null, null, null, 6, '2000-03-23 15:47:14.403000', 6, null, null,
        false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (26, 'Panneau', 'Panneau', null, 26, null, false, null, null, null, 6, '2000-03-23 16:20:27.793000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (27, 'Plage', 'Plage', null, 27, null, false, null, null, null, 6, '2000-03-24 16:19:23.780000', 6, null, null,
        false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (28, 'Débarcadère', 'Débarcadère', null, 28, null, false, null, null, null, 6, '2000-03-24 16:25:30.153000', 6,
        null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (29, 'Passerelle', 'Passerelle', null, 29, 'ThiOuvrageRV', false, null, null, null, 6,
        '2000-04-18 11:02:37.670000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (30, 'Conduite', 'Conduite', null, 30, null, false, null, null, null, 6, '2000-04-18 11:19:03.077000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (31, 'Drapeau', 'Drapeau', null, 31, null, false, null, null, null, 6, '2000-04-18 11:22:28.497000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (32, 'Mât', 'Mât', null, 32, null, false, null, null, null, 6, '2000-04-18 11:22:55.107000', 6, null, null,
        false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (33, 'Pieu', 'Pieu', null, 33, null, false, null, null, null, 6, '2000-04-18 11:30:47.153000', 6, null, null,
        false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (34, 'Place', 'Place, petite zone particulière', null, 34, null, false, null, null, null, 6,
        '2000-04-18 11:36:06.327000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (35, 'Douche', 'Douche', null, 35, null, false, null, null, null, 6, '2000-04-18 11:50:18.043000', 6, null, null,
        false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (36, 'Attelage', 'Attelage', null, 36, null, false, null, null, null, 6, '2001-05-21 10:59:36.263000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (37, 'Estuaire', 'Estuaire', null, 37, null, false, null, null, null, 6, '2001-05-21 11:00:47.530000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (38, 'Tunnel', 'Tunnel', null, 38, 'ThiOuvrageRV', false, null, null, null, 6, '2002-01-23 17:49:14.683000', 6,
        null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (39, 'Pont', 'Pont', null, 39, 'ThiOuvrageRV', false, null, null, null, 6, '2002-02-01 14:44:00.873000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (40, 'Escalier', 'Escalier', null, 40, 'ThiOuvrageRV', false, null, null, null, 6, '2002-02-01 14:45:03.920000',
        6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (41, 'Ouvrage de soutènement', 'Ouvrage de soutènement', null, 41, 'ThiOuvrageRV', false, null, null, null, 6,
        '2002-02-08 14:58:59.403000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (42, 'Passage inférieur', 'Passage inférieur', null, 42, 'ThiOuvrageRV', false, null, null, null, 6,
        '2002-02-08 15:00:44.827000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (43, 'Dalle', 'Dalle', null, 43, 'ThiOuvrageRV', false, null, null, null, 6, '2002-02-08 15:05:32.607000', 6,
        null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (45, 'Mur', 'Mur', null, 45, 'ThiOuvrageRV', false, null, null, null, 6, '2002-02-08 15:13:04.327000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (46, 'Procédé de réclame', 'Procédé de réclame (OSU)', null, 46, 'ProcedeReclame', false, null, null, null, 6,
        '2002-11-08 14:25:10.273000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (47, 'Zone', 'Zone', null, 47, null, false, null, null, null, 6, '2004-04-15 12:12:07.797000', 6, null, null,
        false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (48, 'Infrastructure sportive', 'Infrastructure sportive', null, 48, null, false, null, null, null, 6,
        '2004-04-30 12:49:38.763000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (49, 'Installation sportive', 'Installation sportive', null, 49, 'ThiInstSport', false, null, null, null, 6,
        '2004-07-02 14:26:02.780000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (50, 'Equipement sportif', 'Equipement sportif', null, 50, null, false, null, null, null, 6,
        '2004-07-02 14:26:31.233000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (51, 'Carrefour - Régulation', 'Carrefour - Régulation (Spéc. dans ThiCarReg)', null, 51, null, false, null,
        null, null, 6, '2006-03-01 11:04:14.953000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (52, 'Caméra', 'Caméra - Régulation', null, 52, null, false, null, null, null, 6, '2006-03-01 11:05:06.610000',
        6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (53, 'Parking', 'Parking', null, 53, null, false, null, null, null, 6, '2006-03-01 11:05:26.640000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (54, 'Automate', 'Automate (énergie)', null, 54, 'ThiEnergieAutomate', false, null, null, null, 10958,
        '2006-03-01 11:16:10.223000', 10958, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (55, 'Installation de distribution', 'Installation de distribution (énergie)', null, 55, 'ThiInstDistribution',
        false, null, null, null, 10958, '2006-03-01 11:16:39.677000', 10958, null, null, false, null, null, null,
        'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (56, 'Emplacement de caissette', 'Emplacement de caisssette (OSU)', null, 56, 'ThiCaissetteEmpl', false, null,
        null, null, 10958, '2006-07-28 11:12:49.660000', 10958, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (57, 'Caissette à journaux', 'Caissette à journaux (OSU)', null, 57, 'ThiCaissetteJ', false, null, null, null,
        10958, '2006-07-28 11:13:58.300000', 10958, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (58, 'Parcelle hors Lausanne', 'Parcelle hors Lausanne', null, 58, 'ThiParcelleHorsLs', false, null, null, null,
        6, '2006-10-25 10:49:03.577000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (59, 'Encorbellement', 'Encorbellement', null, 59, 'ThiOuvrageRV', false, null, null, null, 10958,
        '2006-12-21 15:01:11.560000', 10958, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (61, 'Arc boutant', 'Arc boutant', null, 61, 'ThiOuvrageRV', false, null, null, null, 10958,
        '2006-12-21 15:04:13.293000', 10958, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (63, 'Galerie technique', 'Galerie technique', null, 63, 'ThiOuvrageRV', false, null, null, null, 10958,
        '2006-12-21 15:05:41.577000', 10958, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (65, 'Zone d''ancrage', 'Zone d''ancrage', null, 65, null, false, null, null, null, 6,
        '2007-01-18 17:09:59.750000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (66, 'Borne escamotable', 'Borne escamotable', null, 66, null, false, null, null, null, 6,
        '2008-05-14 10:25:20.450000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (67, 'Appareil de surveillance - Régulation', 'Appareil de surveillance - Régulation', null, 67, null, false,
        null, null, null, 6, '2008-05-14 10:26:42.357000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (68, 'Compteur - Régulation', 'Compteur - Régulation', null, 68, null, false, null, null, null, 6,
        '2008-05-14 10:57:41.403000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (69, 'EAU-Adduction', 'EAU-Adduction', null, 69, 'ThiEauAdduction', false, null, null, null, 6,
        '2008-09-02 11:33:16.200000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (70, 'EAU-Sous-adduction', 'EAU-Sous-adduction', null, 70, null, false, null, null, null, 6,
        '2008-09-02 11:33:16.000000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (71, 'EAU-Captage', 'EAU-Captage', null, 71, null, false, null, null, null, 6, '2008-09-02 11:33:16.000000', 6,
        null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (72, 'Emplacement de panneaux', 'Emplacement de panneaux d''affichage (OSU)', null, 72, 'ThiPanneauEmpl', false,
        null, null, null, 10958, '2009-03-03 15:06:17.000000', 10958, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (73, 'Panneau d''affichage', 'Panneau d''affichage (OSU)', null, 73, 'ThiPanneauAff', false, null, null, null,
        10958, '2009-03-03 15:07:34.000000', 10958, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type, icon_path)
VALUES (74, 'Arbre', 'Arbre (SPP)', null, 74, 'ThiArbre', false, null, null, null, 10958, '2009-04-06 09:49:13.000000',
        10958, null, null, false, null, null, null, 'point', '/img/gomarker_tree.png');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (76, 'Sondage géologique', 'Sondage géologique', null, 76, null, false, null, null, null, 12539,
        '2009-10-13 10:33:17.530000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (77, 'Instrument de mesure', 'Instrument de mesure', null, 77, null, false, null, null, null, 14397,
        '2010-03-25 15:34:40.750000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (78, 'Cinéma', 'Cinéma pour plan ville', null, 78, null, false, null, null, null, 7,
        '2010-05-05 15:06:40.000000', 7, null, null, false, null, null, null, 'point');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (79, 'Assainissement - Zone de ramassage ', 'Assainissement - Zone de ramassage ', null, 79, null, false, null,
        null, null, 10958, '2010-05-20 11:41:18.890000', 10958, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (80, 'Poste fixe', 'Poste fixe  (Assainissement)', null, 80, null, false, null, null, null, 10958,
        '2010-08-26 11:41:58.123000', 10958, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (81, 'Déchèterie fixe', 'Déchèterie fixe  (Assainissement)', null, 81, null, false, null, null, null, 10958,
        '2010-08-26 11:43:06.187000', 10958, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (82, 'Déchèterie mobile', 'Déchèterie mobile  (Assainissement)', null, 82, null, false, null, null, null, 10958,
        '2010-08-26 11:43:42.670000', 10958, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (83, 'Site pollué / contaminé', 'Site pollué / contaminé', null, 83, null, false, null, null, null, 6,
        '2011-03-22 08:38:33.327000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (84, 'Place de jeux', 'Place de jeux (SPP)', null, 84, null, false, null, null, null, 10958,
        '2011-05-06 11:23:47.967000', 10958, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (85, 'Ecole', 'Ecole', null, 85, null, false, null, null, null, 12539, '2011-06-08 13:59:34.107000', 6, null,
        null, false, null, null, null, 'point');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (86, 'Ilot urbanistique', 'Ilot urbanistique', null, 86, null, false, null, null, null, 6,
        '2012-01-13 13:06:27.013000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (87, 'Centre de vie enfantine', 'Centre de vie enfantine', null, 87, null, false, null, null, null, 7,
        '2012-02-14 10:49:15.530000', 7, null, null, false, null, null, null, 'point');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (88, 'Pro Infirmis', 'Pro Infirmis', null, 88, null, false, null, null, null, 7, '2012-05-02 09:40:53.060000', 7,
        null, null, false, null, null, null, 'point');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (89, 'Logement subventionné', 'Logement subventionné pour Internet', null, 89, null, false, null, null, null, 7,
        '2012-09-20 08:45:15.437000', 7, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (90, 'Terminal OST', 'Terminal OST - Régulation', null, 90, null, false, null, null, null, 6,
        '2013-03-22 14:22:19.733000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (91, 'Point fixe planimétrique', 'Point fixe planinétrique', null, 91, 'ThiPointFixe', false, null, null, null,
        10958, '2013-06-26 11:45:40.217000', 10958, null, null, false, null, null, null, 'point');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (92, 'Point fixe altimétrique', 'Point fixe altimétrique', null, 92, 'ThiPointFixe', false, null, null, null,
        10958, '2013-06-26 11:48:57.577000', 10958, null, null, false, null, null, null, 'point');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (94, 'Toiture végétalisée', 'Toiture végétalisée (SPADOM)', null, 94, 'ThiToitureVege', false, null, null, null,
        10958, '2014-01-15 14:27:54.607000', 10958, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (95, 'Emplacement SPADOM', 'Emplacement SPADOM', null, 95, null, false, null, null, null, 10958,
        '2015-03-13 10:45:23.603000', 10958, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (96, 'Sonde géothermique', 'Sonde géothermique (SCC)', null, 96, 'ThiSondageGeoTherm', false, null, null, null,
        10958, '2015-11-09 08:55:49.043000', 10958, null, null, false, null, null, null, 'point');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (97, 'EAU-Site', 'EAU-Site', null, 97, null, false, null, null, null, 6, '2016-06-21 12:47:46.723000', 6, null,
        null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (98, 'Emprise de chantier', 'Emprise de chantier', null, 98, null, false, null, null, null, 6,
        '2016-09-05 12:50:41.097000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (99, 'Emprise de projet', 'Emprise de projet', null, 99, null, false, null, null, null, 7,
        '2016-09-06 10:33:32.193000', 7, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (100, 'Cours d''eau', 'Cours d''eau', null, 100, null, false, null, null, null, 6, '2017-04-24 09:56:07.867000',
        6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (101, 'Servitude', 'Servitude', null, 101, 'thi_servitude', false, null, null, null, 6,
        '2018-01-29 10:40:16.537000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (102, 'Groupement d''arbres', 'Groupement d''arbres (UGA)', null, 102, 'ThiGrpArbre', false, null, null, null,
        10958, '2020-02-24 09:43:01.820000', 10958, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (103, 'Polygone utilisateur', 'Polygone utilisateur pour zone de recherche', null, 103, null, false, null, null,
        null, 6, '2020-03-11 10:35:16.997000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (104, 'SPADOM - Ouvrages - Escaliers', 'Ouvrages SPADOM QGIS Escaliers', null, 104, null, false, null, null,
        null, 7, '2020-12-16 15:55:07.103000', 7, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (105, 'SPADOM - Ouvrages - Murs de soutènement', 'Ouvrages SPADOM QGIS Mur de soutènement', null, 105, null,
        false, null, null, null, 7, '2020-12-22 15:51:46.997000', 7, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (106, 'SPADOM - Ouvrages - Garde-corps', 'Ouvrages SPADOM QGIS Garde-corps', null, 106, null, false, null, null,
        null, 7, '2020-12-22 17:45:35.687000', 7, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (107, 'SPADOM - Ouvrages - Passerelle', 'Ouvrages SPADOM QGIS Passerelle', null, 107, null, false, null, null,
        null, 7, '2021-03-08 15:54:01.947000', 7, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (108, 'Terrains laissés en jouissance', 'Terrains laissés en jouissance (bien-plaire)', null, 108, null, false,
        null, null, null, 6, '2021-09-07 08:25:55.200000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (109, 'Ouvrage de rétention', 'Ouvrage de rétention', null, 109, 'thi_ouvrage_retention', false, null, null,
        null, 6, '2021-09-14 10:57:41.717000', 6, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (110, 'Antenne de téléphonie mobile', 'Antenne de téléphonie mobile', null, 110, null, false, null, null, null,
        10958, '2022-04-20 00:00:00.000000', 10958, null, null, false, null, null, null, 'point');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (111, 'SPADOM - Ouvrages - Pont', 'Ouvrage SPADOM QGIS Pont', null, 111, null, false, null, null, null, 7,
        '2022-04-25 15:05:35.273000', 7, null, null, false, null, null, null, 'bbox');
INSERT INTO template_4_your_project_name_db_schema.type_template_4_your_project_name (id, name, description, comment, external_id, table_name, inactivated, inactivated_time,
                                 inactivated_by, inactivated_reason, managed_by, _created_at, _created_by,
                                 _last_modified_at, _last_modified_by, _deleted, _deleted_at, _deleted_by,
                                 more_data_schema, geometry_type)
VALUES (112, 'SPADOM - Ouvrages - Oeuvre d''art', 'Ouvrage SPADOM QGIS Oeuvre d''art', null, 112, null, false, null,
        null, null, 7, '2022-04-25 16:03:03.180000', 7, null, null, false, null, null, null, 'bbox');






UPDATE template_4_your_project_name_db_schema.type_template_4_your_project_name
SET text_search = to_tsvector('french',
                              unaccent(name) ||
                              ' ' || coalesce(unaccent(description), ' ') ||
                              ' ' || coalesce(unaccent(comment), ' '))
WHERE text_search IS NULL;

create index type_template_4_your_project_name_text_search_index
    on template_4_your_project_name_db_schema.type_template_4_your_project_name using gin (text_search);

SELECT setval('template_4_your_project_name_db_schema.type_template_4_your_project_name_id_seq', max(id)) FROM template_4_your_project_name_db_schema.type_template_4_your_project_name;
