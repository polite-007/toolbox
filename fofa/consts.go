package fofa

// SearchFieldsFree 免费版可用的查询字段（34个）
const SearchFieldsFree = "ip,port,protocol,country,country_name,region,city," +
	"longitude,latitude,asn,org,host,domain,os,server," +
	"icp,title,jarm,header,banner,cert,base_protocol,link," +
	"cert.issuer.org,cert.issuer.cn,cert.subject.org,cert.subject.cn," +
	"tls.ja3s,tls.version,cert.sn,cert.not_before,cert.not_after," +
	"cert.domain,status_code"

// SearchFieldsPersonal 个人版及以上可用的查询字段（37个，含免费版）
const SearchFieldsPersonal = SearchFieldsFree +
	",header_hash,banner_hash,banner_fid"

// SearchFieldsProfessional 专业版及以上可用的查询字段（41个，含个人版）
const SearchFieldsProfessional = SearchFieldsPersonal +
	",cname,lastupdatetime,product,product_category"

// SearchFieldsBusiness 商业版及以上可用的查询字段（48个，含专业版）
const SearchFieldsBusiness = SearchFieldsProfessional +
	",product.version,icon_hash,cert.is_valid,cname_domain," +
	"body,cert.is_match,cert.is_equal"

// SearchFieldsEnterprise 企业版可用的查询字段（51个，含商业版）
const SearchFieldsEnterprise = SearchFieldsBusiness +
	",icon,fid,structinfo"
