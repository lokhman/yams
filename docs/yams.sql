CREATE EXTENSION "pgcrypto";
------------------------------------------------------------------------------------------------------------------------
CREATE TYPE adapter AS ENUM('lua');
CREATE TYPE role AS ENUM('developer', 'manager', 'admin');
------------------------------------------------------------------------------------------------------------------------
CREATE TABLE profiles
(
  id            serial                              NOT NULL
    CONSTRAINT profiles_pkey
    PRIMARY KEY,
  name          varchar(72)                         NOT NULL,
  backend       varchar(128),
  hosts         varchar(128)[]                      NOT NULL,
  is_debug      boolean DEFAULT TRUE                NOT NULL,
  vars_lifetime integer DEFAULT 86400               NOT NULL,
  created_at    timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL
);
------------------------------------------------------------------------------------------------------------------------
CREATE TABLE routes
(
  id         serial                           NOT NULL
    CONSTRAINT routes_pkey
    PRIMARY KEY,
  uuid       uuid DEFAULT gen_random_uuid()   NOT NULL,
  profile_id integer                          NOT NULL
    CONSTRAINT routes_profiles_id_fk
    REFERENCES profiles
    ON DELETE CASCADE,
  methods    varchar(16)[]                    NOT NULL,
  path       varchar(2048)                    NOT NULL,
  path_re    varchar(4096)                    NOT NULL,
  path_args  varchar(2048)[]                  NOT NULL,
  script     bytea DEFAULT '\x'::bytea        NOT NULL,
  position   integer                          NOT NULL,
  timeout    integer DEFAULT 60               NOT NULL,
  hint       varchar(255),
  adapter    adapter DEFAULT 'lua'::adapter   NOT NULL,
  is_enabled boolean DEFAULT TRUE             NOT NULL
);

CREATE UNIQUE INDEX routes_uuid_uindex ON routes (uuid);
------------------------------------------------------------------------------------------------------------------------
CREATE TABLE assets
(
  id         serial                                NOT NULL
    CONSTRAINT assets_pkey
    PRIMARY KEY,
  profile_id integer                               NOT NULL
    CONSTRAINT assets_profiles_id_fk
    REFERENCES profiles
    ON DELETE CASCADE,
  path       varchar(72) DEFAULT gen_random_uuid() NOT NULL,
  data       bytea                                 NOT NULL,
  mime_type  varchar(255)                          NOT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP   NOT NULL
);

CREATE UNIQUE INDEX assets_profile_id_path_uindex ON assets (profile_id, path);
------------------------------------------------------------------------------------------------------------------------
CREATE TABLE storage
(
  profile_id integer      NOT NULL
    CONSTRAINT storage_profiles_id_fk
    REFERENCES profiles
    ON DELETE CASCADE,
  sid        varchar(24),
  key        varchar(255) NOT NULL,
  value      json         NOT NULL,
  expires_at timestamp    NOT NULL,
  updated_at timestamp    NOT NULL
);

CREATE UNIQUE INDEX storage_profile_id_sid_key_uindex ON storage (profile_id, COALESCE(sid, ''::varchar), key);
------------------------------------------------------------------------------------------------------------------------
CREATE TABLE users
(
  id           serial                              NOT NULL
    CONSTRAINT users_pkey
    PRIMARY KEY,
  username     varchar(32)                         NOT NULL,
  password     char(72)                            NOT NULL,
  role         role                                NOT NULL,
  created_at   timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
  last_auth_at timestamp,
  updated_at   timestamp                           NOT NULL
);

CREATE UNIQUE INDEX users_username_uindex ON users (username);
------------------------------------------------------------------------------------------------------------------------
CREATE TABLE acl
(
  user_id    integer NOT NULL
    CONSTRAINT acl_users_id_fk
    REFERENCES users
    ON DELETE CASCADE,
  profile_id integer NOT NULL
    CONSTRAINT acl_profiles_id_fk
    REFERENCES profiles
    ON DELETE CASCADE,
  CONSTRAINT acl_pkey
  PRIMARY KEY (user_id, profile_id)
);
------------------------------------------------------------------------------------------------------------------------
CREATE FUNCTION yams_profiles_robot() RETURNS TRIGGER AS $$
BEGIN
  NEW.hosts = ARRAY(SELECT DISTINCT regexp_replace(lower(unnest(NEW.hosts)), ':(?:80|443)$', '') AS x ORDER BY x);
  NEW.backend = rtrim(regexp_replace(lower(NEW.backend), '^((?:[^\/]*\/){3}).*', '\1'), '/');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER profiles_robot BEFORE INSERT OR UPDATE ON profiles
  FOR EACH ROW EXECUTE PROCEDURE yams_profiles_robot();
------------------------------------------------------------------------------------------------------------------------
CREATE FUNCTION yams_routes_robot() RETURNS TRIGGER AS $$
BEGIN
  NEW.methods = ARRAY(SELECT DISTINCT upper(unnest(NEW.methods)) AS x ORDER BY x);
  NEW.path_re = '^' || regexp_replace(
      regexp_replace(NEW.path, '([!$()*+.:<=>?[\\\]^{|}-])', '\\\1', 'g'), '\\\{\w+\\\}', '(.*)', 'g') || '$';
  NEW.path_args = ARRAY(SELECT unnest(regexp_matches(NEW.path, '\{(\w+)\}', 'g')));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER routes_robot BEFORE INSERT OR UPDATE ON routes
  FOR EACH ROW EXECUTE PROCEDURE yams_routes_robot();
------------------------------------------------------------------------------------------------------------------------
CREATE FUNCTION yams_routes_position() RETURNS TRIGGER AS $$
BEGIN
  IF tg_op = 'INSERT' THEN
    NEW.position = (SELECT count(*) FROM routes WHERE profile_id = NEW.profile_id);
    RETURN NEW;
  ELSEIF tg_op = 'UPDATE' THEN
    IF OLD.profile_id <> NEW.profile_id THEN
      NEW.profile_id = OLD.profile_id;
    END IF;

    NEW.position = least(greatest(0, NEW.position),
        (SELECT COALESCE(max(position), 0) FROM routes WHERE profile_id = NEW.profile_id));

    IF OLD.position < NEW.position THEN
      UPDATE routes SET position = position - 1
          WHERE profile_id = NEW.profile_id AND position > OLD.position AND position <= NEW.position;
    ELSEIF OLD.position > NEW.position THEN
      UPDATE routes SET position = position + 1
          WHERE profile_id = NEW.profile_id AND position >= NEW.position AND position < OLD.position;
    END IF;
    RETURN NEW;
  ELSEIF tg_op = 'DELETE' THEN
    UPDATE routes SET position = position - 1 WHERE profile_id = OLD.profile_id AND position > OLD.position;
    RETURN OLD;
  END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER routes_position BEFORE INSERT OR UPDATE OR DELETE ON routes
  FOR EACH ROW WHEN (pg_trigger_depth() = 0) EXECUTE PROCEDURE yams_routes_position();
------------------------------------------------------------------------------------------------------------------------
CREATE FUNCTION yams_storage_robot() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER storage_robot BEFORE INSERT OR UPDATE ON storage
  FOR EACH ROW EXECUTE PROCEDURE yams_storage_robot();
------------------------------------------------------------------------------------------------------------------------
CREATE FUNCTION yams_storage_recycle() RETURNS TRIGGER AS $$
BEGIN
  DELETE FROM storage WHERE expires_at <= now();
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER storage_recycle AFTER INSERT OR UPDATE ON storage
  FOR STATEMENT EXECUTE PROCEDURE yams_storage_recycle();
------------------------------------------------------------------------------------------------------------------------
CREATE FUNCTION yams_users_robot() RETURNS TRIGGER AS $$
BEGIN
  NEW.username = lower(NEW.username);
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_robot BEFORE INSERT OR UPDATE ON users
  FOR EACH ROW EXECUTE PROCEDURE yams_users_robot();
------------------------------------------------------------------------------------------------------------------------
CREATE FUNCTION yams_assets_robot() RETURNS TRIGGER AS $$
BEGIN
  NEW.path = lower(NEW.path);
  NEW.mime_type = lower(NEW.mime_type);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER assets_robot BEFORE INSERT OR UPDATE ON assets
  FOR EACH ROW EXECUTE PROCEDURE yams_assets_robot();
------------------------------------------------------------------------------------------------------------------------
INSERT INTO users (username, role, password) VALUES ('admin', 'admin', crypt('admin', gen_salt('bf')));
