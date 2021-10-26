create table currencies
(
	id bigserial not null
		constraint currencies_pk
			primary key,
	code text not null,
	multiplier bigint not null,
	name text not null,
    sign text,
    description text
);

alter table currencies owner to postgres;

create unique index currencies_code_uindex
	on currencies (code);

create unique index currencies_id_uindex
	on currencies (id);

create table security_types
(
	id bigserial not null
		constraint security_types_pk
			primary key,
	name text not null
);

alter table security_types owner to postgres;

create unique index security_types_id_uindex
	on security_types (id);

create unique index security_types_name_uindex
	on security_types (name);

create table trade_states
(
	id bigserial not null
		constraint trade_states_pk
			primary key,
	name text not null
);

alter table trade_states owner to postgres;

create unique index trade_states_id_uindex
	on trade_states (id);

create unique index trade_states_name_uindex
	on trade_states (name);

create table data_sources
(
	id bigserial not null
		constraint data_sources_pk
			primary key,
	name text not null
);

alter table data_sources owner to postgres;

create table security_info
(
	id bigserial not null
		constraint security_info_pk
			primary key,
	ticker text not null,
	figi text not null,
	trading_status bigint not null
		constraint security_info_trade_states__fk
			references trade_states
				on update restrict on delete restrict,
	currency bigint not null
		constraint security_info_currencies_id_fk
			references currencies,
	lot_size bigint not null,
	price_increment bigint not null,
	type bigint not null
		constraint security_info_security_types__fk
			references security_types,
	data_source bigint not null
		constraint security_info_data_source_fk
			references data_sources
				on delete restrict
);

alter table security_info owner to postgres;

create unique index security_info_currency_uindex
	on security_info (currency);

create unique index security_info_id_uindex
	on security_info (id);

create table candles
(
	id bigserial not null
		constraint candles_pk
			primary key,
	security bigint not null
		constraint security_fk
			references security_info
				on delete cascade,
	datetime timestamp with time zone not null,
	open_price bigint not null,
	close_price bigint not null,
	high_price bigint not null,
	low_price bigint not null
);

alter table candles owner to postgres;

create unique index candles_id_uindex
	on candles (id);

create unique index data_sources_id_uindex
	on data_sources (id);

create unique index data_sources_name_uindex
	on data_sources (name);

create table order_book_entries
(
	id bigserial not null
		constraint order_book_security_pk
			primary key,
	security bigint not null
		constraint order_book_security__security_info__fk
			references security_info
				on delete cascade,
	datetime timestamp with time zone not null
);

alter table order_book_entries owner to postgres;

create unique index order_book_security_id_uindex
	on order_book_entries (id);

create table order_book_price
(
	id bigserial not null
		constraint order_book_price_pk
			primary key,
	price bigint not null,
	count bigint not null
);

alter table order_book_price owner to postgres;

create unique index order_book_price_id_uindex
	on order_book_price (id);

create table order_book
(
	entry_id bigint not null
		constraint order_book_entry___fk
			references order_book_entries
				on delete cascade,
	price_id bigint
		constraint order_book_price__fk
			references order_book_price
				on delete cascade
);

alter table order_book owner to postgres;

