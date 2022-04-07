
CREATE DOMAIN public.xid AS TEXT CHECK (VALUE ~ '^[a-z0-9]{20}$');

CREATE SEQUENCE public.xid_serial MINVALUE 1 MAXVALUE 16777215 CYCLE ; --  ((255<<16) + (255<<8) + 255))

SELECT setval('xid_serial', (random() * 16777215)::INT);  --  ((255<<16) + (255<<8) + 255))

CREATE OR REPLACE FUNCTION public.xid_encode(_id int[])
    RETURNS public.xid
    LANGUAGE plpgsql
AS
$$
DECLARE
    _encoding TEXT[] = '{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p, q, r, s, t, u, v}';
BEGIN
    RETURN _encoding[1+(_id[1] >> 3)]
       || _encoding[1+((_id[2]>>6)&31|(_id[1]<<2)&31)]
       || _encoding[1+((_id[2]>>1)&31)]
       || _encoding[1+((_id[3]>>4)&31|(_id[2]<<4)&31)]
       || _encoding[1+(_id[4]>>7|(_id[3]<<1)&31)]
       || _encoding[1+((_id[4]>>2)&31)]
       || _encoding[1+(_id[5]>>5|(_id[4]<<3)&31)]
       || _encoding[1+(_id[5]&31)]
       || _encoding[1+(_id[6]>>3)]
       || _encoding[1+((_id[7]>>6)&31|(_id[6]<<2)&31)]
       || _encoding[1+((_id[7]>>1)&31)]
       || _encoding[1+((_id[8]>>4)&31|(_id[7]<<4)&31)]
       || _encoding[1+(_id[9]>>7|(_id[8]<<1)&31)]
       || _encoding[1+((_id[9]>>2)&31)]
       || _encoding[1+((_id[10]>>5)|(_id[9]<<3)&31)]
       || _encoding[1+(_id[10]&31)]
       || _encoding[1+(_id[11]>>3)]
       || _encoding[1+((_id[12]>>6)&31|(_id[11]<<2)&31)]
       || _encoding[1+((_id[12]>>1)&31)]
       || _encoding[1+((_id[12]<<4)&31)];
END;
$$;

CREATE OR REPLACE FUNCTION public.xid_decode(_xid public.xid)
    RETURNS int[]
    LANGUAGE plpgsql
AS
$$
DECLARE
    _dec int[] = '{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}';

BEGIN
    return ARRAY[
            ((_dec[ascii(substring(_xid,1))]<<3) | (_dec[ascii(substring(_xid, 2))]>>2)) & 255,
            ((_dec[ascii(substring(_xid,2))]<<6) | (_dec[ascii(substring(_xid,3))]<<1) | (_dec[ascii(substring(_xid,4))]>>4)) & 255,
            ((_dec[ascii(substring(_xid,4))]<<4) | (_dec[ascii(substring(_xid,5))]>>1)) & 255,
            ((_dec[ascii(substring(_xid,5))]<<7) | (_dec[ascii(substring(_xid,6))]<<2) | (_dec[ascii(substring(_xid,7))]>>3)) & 255,
            ((_dec[ascii(substring(_xid,7))]<<5) | (_dec[ascii(substring(_xid,8))])) & 255,
            ((_dec[ascii(substring(_xid,9))]<<3) | (_dec[ascii(substring(_xid,10))]>>2)) & 255,
            ((_dec[ascii(substring(_xid,10))]<<6) | (_dec[ascii(substring(_xid,11))]<<1) | (_dec[ascii(substring(_xid,12))]>>4)) & 255,
            ((_dec[ascii(substring(_xid,12))]<<4) | (_dec[ascii(substring(_xid,13))]>>1)) & 255,
            ((_dec[ascii(substring(_xid,13))]<<7) | (_dec[ascii(substring(_xid,14))]<<2) | (_dec[ascii(substring(_xid,15))]>>3)) & 255,
            ((_dec[ascii(substring(_xid,15))]<<5) | (_dec[ascii(substring(_xid,16))])) & 255,
            ((_dec[ascii(substring(_xid,17))]<<3) | (_dec[ascii(substring(_xid,18))]>>2)) & 255,
            ((_dec[ascii(substring(_xid,18))]<<6) | (_dec[ascii(substring(_xid,19))]<<1) | (_dec[ascii(substring(_xid,20))]>>4)) & 255
        ];
END;
$$;

CREATE OR REPLACE FUNCTION xid(_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP)
    RETURNS public.xid
    LANGUAGE plpgsql
AS
$$
DECLARE
    _bytes        BYTEA;
    _id     int[];
BEGIN

    _bytes := int8send(floor(EXTRACT(epoch FROM _at))::BIGINT);
    _id := ARRAY [get_byte(_bytes, 4), get_byte(_bytes, 5), get_byte(_bytes, 6) , get_byte(_bytes, 7)];

    _bytes := int8send((SELECT system_identifier  FROM pg_control_system()));
    _id := _id || ARRAY [ get_byte(_bytes, 5) , get_byte(_bytes, 6) , get_byte(_bytes, 7)];

    _bytes := int4send(pg_backend_pid());
    _id := _id || ARRAY [ get_byte(_bytes, 2), get_byte(_bytes, 3)];

    _bytes := int4send(nextval('xid_serial')::INT);
    _id := _id || ARRAY [get_byte(_bytes, 1), get_byte(_bytes, 2), get_byte(_bytes, 3)];

    return public.xid_encode(_id)
   ;
END;
$$;

CREATE OR REPLACE FUNCTION public.xid_time(_xid public.xid)
    RETURNS TIMESTAMPTZ
    LANGUAGE plpgsql
AS
$$
DECLARE
    _id int[];
BEGIN
    _id := public.xid_decode(_xid);
    return to_timestamp( (_id[1] << 24)::BIGINT + (_id[2] << 16) + (_id[3] << 8) + (_id[4]) );
END;
$$;

CREATE OR REPLACE FUNCTION public.xid_machine(_xid public.xid)
    RETURNS INT[]
    LANGUAGE plpgsql
AS
$$
DECLARE
    _id int[];
BEGIN
    _id := public.xid_decode(_xid);
    return  ARRAY [_id[5], _id[6], _id[7]] ;
END;
$$;

CREATE OR REPLACE FUNCTION public.xid_pid(_xid public.xid)
    RETURNS INT
    LANGUAGE plpgsql
AS
$$
DECLARE
    _id int[];
BEGIN
    _id := public.xid_decode(_xid);
    return  (_id[8] << 8) + (_id[9]) ;
END;
$$;

CREATE OR REPLACE FUNCTION public.xid_counter(_xid public.xid)
    RETURNS INT
    LANGUAGE plpgsql
AS
$$
DECLARE
    _id int[];
BEGIN
    _id := public.xid_decode(_xid);
    return  (_id[10] << 16) + (_id[11] << 8) + (_id[12]) ;
END;
$$;
