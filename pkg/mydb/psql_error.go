package mydb

import (
	"encoding/json"
	"strings"
)

type PSQLError string

const (
	UnknownPSQLError = "?????"
	// Class 00 — Successful Completion
	SuccessfulCompletion PSQLError = "00000"
	// Class 01 — Warning
	Warning                          PSQLError = "01000"
	DynamicResultSetsReturned        PSQLError = "0100C"
	ImplicitZeroBitPadding           PSQLError = "01008"
	NullValueEliminatedInSetFunction PSQLError = "01003"
	PrivilegeNotGranted              PSQLError = "01007"
	PrivilegeNotRevoked              PSQLError = "01006"
	WarningStringDataRightTruncation PSQLError = "01004"
	DeprecatedFeature                PSQLError = "01P01"
	// Class 02 — No Data (this is also a warning class per the SQL standard)
	NoData                                PSQLError = "02000"
	NoAdditionalDynamicResultSetsReturned PSQLError = "02001"
	// Class 03 — SQL Statement Not Yet Complete
	SqlStatementNotYetComplete PSQLError = "03000"
	// Class 08 — Connection Exception
	ConnectionException                           PSQLError = "08000"
	ConnectionDoesNotExist                        PSQLError = "08003"
	ConnectionFailure                             PSQLError = "08006"
	SqlclientUnableToEstablishSqlconnection       PSQLError = "08001"
	SqlserverRejectedEstablishmentOfSqlconnection PSQLError = "08004"
	TransactionResolutionUnknown                  PSQLError = "08007"
	ProtocolViolation                             PSQLError = "08P01"
	// Class 09 — Triggered Action Exception
	TriggeredActionException PSQLError = "09000"
	// Class 0A — Feature Not Supported
	FeatureNotSupported PSQLError = "0A000"
	// Class 0B — Invalid Transaction Initiation
	InvalidTransactionInitiation PSQLError = "0B000"
	// Class 0F — Locator Exception
	LocatorException            PSQLError = "0F000"
	InvalidLocatorSpecification PSQLError = "0F001"
	// Class 0L — Invalid Grantor
	InvalidGrantor        PSQLError = "0L000"
	InvalidGrantOperation PSQLError = "0LP01"
	// Class 0P — Invalid Role Specification
	InvalidRoleSpecification PSQLError = "0P000"
	// Class 0Z — Diagnostics Exception
	DiagnosticsException                           PSQLError = "0Z000"
	StackedDiagnosticsAccessedWithoutActiveHandler PSQLError = "0Z002"
	// Class 20 — Case Not Found
	CaseNotFound PSQLError = "20000"
	// Class 21 — Cardinality Violation
	CardinalityViolation PSQLError = "21000"
	// Class 22 — Data Exception
	DataException                         PSQLError = "22000"
	ArraySubscriptError                   PSQLError = "2202E"
	CharacterNotInRepertoire              PSQLError = "22021"
	DatetimeFieldOverflow                 PSQLError = "22008"
	DivisionByZero                        PSQLError = "22012"
	ErrorInAssignment                     PSQLError = "22005"
	EscapeCharacterConflict               PSQLError = "2200B"
	IndicatorOverflow                     PSQLError = "22022"
	IntervalFieldOverflow                 PSQLError = "22015"
	InvalidArgumentForLogarithm           PSQLError = "2201E"
	InvalidArgumentForNtileFunction       PSQLError = "22014"
	InvalidArgumentForNthValueFunction    PSQLError = "22016"
	InvalidArgumentForPowerFunction       PSQLError = "2201F"
	InvalidArgumentForWidthBucketFunction PSQLError = "2201G"
	InvalidCharacterValueForCast          PSQLError = "22018"
	InvalidDatetimeFormat                 PSQLError = "22007"
	InvalidEscapeCharacter                PSQLError = "22019"
	InvalidEscapeOctet                    PSQLError = "2200D"
	InvalidEscapeSequence                 PSQLError = "22025"
	NonstandardUseOfEscapeCharacter       PSQLError = "22P06"
	InvalidIndicatorParameterValue        PSQLError = "22010"
	InvalidParameterValue                 PSQLError = "22023"
	InvalidRegularExpression              PSQLError = "2201B"
	InvalidRowCountInLimitClause          PSQLError = "2201W"
	InvalidRowCountInResultOffsetClause   PSQLError = "2201X"
	InvalidTimeZoneDisplacementValue      PSQLError = "22009"
	InvalidUseOfEscapeCharacter           PSQLError = "2200C"
	MostSpecificTypeMismatch              PSQLError = "2200G"
	NullValueNotAllowed                   PSQLError = "22004"
	NullValueNoIndicatorParameter         PSQLError = "22002"
	NumericValueOutOfRange                PSQLError = "22003"
	StringDataLengthMismatch              PSQLError = "22026"
	StringDataRightTruncation             PSQLError = "22001"
	SubstringError                        PSQLError = "22011"
	TrimError                             PSQLError = "22027"
	UnterminatedCString                   PSQLError = "22024"
	ZeroLengthCharacterString             PSQLError = "2200F"
	FloatingPointException                PSQLError = "22P01"
	InvalidTextRepresentation             PSQLError = "22P02"
	InvalidBinaryRepresentation           PSQLError = "22P03"
	BadCopyFileFormat                     PSQLError = "22P04"
	UntranslatableCharacter               PSQLError = "22P05"
	NotAnXmlDocument                      PSQLError = "2200L"
	InvalidXmlDocument                    PSQLError = "2200M"
	InvalidXmlContent                     PSQLError = "2200N"
	InvalidXmlComment                     PSQLError = "2200S"
	InvalidXmlProcessingInstruction       PSQLError = "2200T"
	// Class 23 — Integrity Constraint Violation
	IntegrityConstraintViolation PSQLError = "23000"
	RestrictViolation            PSQLError = "23001"
	NotNullViolation             PSQLError = "23502"
	ForeignKeyViolation          PSQLError = "23503"
	UniqueViolation              PSQLError = "23505"
	CheckViolation               PSQLError = "23514"
	ExclusionViolation           PSQLError = "23P01"
	// Class 24 — Invalid Cursor State
	InvalidCursorState PSQLError = "24000"
	// Class 25 — Invalid Transaction State
	InvalidTransactionState                         PSQLError = "25000"
	ActiveSqlTransaction                            PSQLError = "25001"
	BranchTransactionAlreadyActive                  PSQLError = "25002"
	HeldCursorRequiresSameIsolationLevel            PSQLError = "25008"
	InappropriateAccessModeForBranchTransaction     PSQLError = "25003"
	InappropriateIsolationLevelForBranchTransaction PSQLError = "25004"
	NoActiveSqlTransactionForBranchTransaction      PSQLError = "25005"
	ReadOnlySqlTransaction                          PSQLError = "25006"
	SchemaAndDataStatementMixingNotSupported        PSQLError = "25007"
	NoActiveSqlTransaction                          PSQLError = "25P01"
	InFailedSqlTransaction                          PSQLError = "25P02"
	// Class 26 — Invalid SQL Statement Name
	InvalidSqlStatementName PSQLError = "26000"
	// Class 27 — Triggered Data Change Violation
	TriggeredDataChangeViolation PSQLError = "27000"
	// Class 28 — Invalid Authorization Specification
	InvalidAuthorizationSpecification PSQLError = "28000"
	InvalidPassword                   PSQLError = "28P01"
	// Class 2B — Dependent Privilege Descriptors Still Exist
	DependentPrivilegeDescriptorsStillExist PSQLError = "2B000"
	DependentObjectsStillExist              PSQLError = "2BP01"
	// Class 2D — Invalid Transaction Termination
	InvalidTransactionTermination PSQLError = "2D000"
	// Class 2F — SQL Routine Exception
	SqlRoutineException               PSQLError = "2F000"
	FunctionExecutedNoReturnStatement PSQLError = "2F005"
	ModifyingSqlDataNotPermitted      PSQLError = "2F002"
	ProhibitedSqlStatementAttempted   PSQLError = "2F003"
	ReadingSqlDataNotPermitted        PSQLError = "2F004"
	// Class 34 — Invalid Cursor Name
	InvalidCursorName PSQLError = "34000"
	// Class 38 — External Routine Exception
	ExternalRoutineException           PSQLError = "38000"
	ContainingSqlNotPermitted          PSQLError = "38001"
	EreModifyingSqlDataNotPermitted    PSQLError = "38002"
	EreProhibitedSqlStatementAttempted PSQLError = "38003"
	EreReadingSqlDataNotPermitted      PSQLError = "38004"
	// Class 39 — External Routine Invocation Exception
	ExternalRoutineInvocationException PSQLError = "39000"
	InvalidSqlstateReturned            PSQLError = "39001"
	ErieNullValueNotAllowed            PSQLError = "39004"
	TriggerProtocolViolated            PSQLError = "39P01"
	SrfProtocolViolated                PSQLError = "39P02"
	// Class 3B — Savepoint Exception
	SavepointException            PSQLError = "3B000"
	InvalidSavepointSpecification PSQLError = "3B001"
	// Class 3D — Invalid Catalog Name
	InvalidCatalogName PSQLError = "3D000"
	// Class 3F — Invalid Schema Name
	InvalidSchemaName PSQLError = "3F000"
	// Class 40 — Transaction Rollback
	TransactionRollback                     PSQLError = "40000"
	TransactionIntegrityConstraintViolation PSQLError = "40002"
	SerializationFailure                    PSQLError = "40001"
	StatementCompletionUnknown              PSQLError = "40003"
	DeadlockDetected                        PSQLError = "40P01"
	// Class 42 — Syntax Error or Access Rule Violation
	SyntaxErrorOrAccessRuleViolation   PSQLError = "42000"
	SyntaxError                        PSQLError = "42601"
	InsufficientPrivilege              PSQLError = "42501"
	CannotCoerce                       PSQLError = "42846"
	GroupingError                      PSQLError = "42803"
	WindowingError                     PSQLError = "42P20"
	InvalidRecursion                   PSQLError = "42P19"
	InvalidForeignKey                  PSQLError = "42830"
	InvalidName                        PSQLError = "42602"
	NameTooLong                        PSQLError = "42622"
	ReservedName                       PSQLError = "42939"
	DatatypeMismatch                   PSQLError = "42804"
	IndeterminateDatatype              PSQLError = "42P18"
	CollationMismatch                  PSQLError = "42P21"
	IndeterminateCollation             PSQLError = "42P22"
	WrongObjectType                    PSQLError = "42809"
	UndefinedColumn                    PSQLError = "42703"
	UndefinedFunction                  PSQLError = "42883"
	UndefinedTable                     PSQLError = "42P01"
	UndefinedParameter                 PSQLError = "42P02"
	UndefinedObject                    PSQLError = "42704"
	DuplicateColumn                    PSQLError = "42701"
	DuplicateCursor                    PSQLError = "42P03"
	DuplicateDatabase                  PSQLError = "42P04"
	DuplicateFunction                  PSQLError = "42723"
	DuplicatePreparedStatement         PSQLError = "42P05"
	DuplicateSchema                    PSQLError = "42P06"
	DuplicateTable                     PSQLError = "42P07"
	DuplicateAlias                     PSQLError = "42712"
	DuplicateObject                    PSQLError = "42710"
	AmbiguousColumn                    PSQLError = "42702"
	AmbiguousFunction                  PSQLError = "42725"
	AmbiguousParameter                 PSQLError = "42P08"
	AmbiguousAlias                     PSQLError = "42P09"
	InvalidColumnReference             PSQLError = "42P10"
	InvalidColumnDefinition            PSQLError = "42611"
	InvalidCursorDefinition            PSQLError = "42P11"
	InvalidDatabaseDefinition          PSQLError = "42P12"
	InvalidFunctionDefinition          PSQLError = "42P13"
	InvalidPreparedStatementDefinition PSQLError = "42P14"
	InvalidSchemaDefinition            PSQLError = "42P15"
	InvalidTableDefinition             PSQLError = "42P16"
	InvalidObjectDefinition            PSQLError = "42P17"
	// Class 44 — WITH CHECK OPTION Violation
	WithCheckOptionViolation PSQLError = "44000"
	// Class 53 — Insufficient Resources
	InsufficientResources      PSQLError = "53000"
	DiskFull                   PSQLError = "53100"
	OutOfMemory                PSQLError = "53200"
	TooManyConnections         PSQLError = "53300"
	ConfigurationLimitExceeded PSQLError = "53400"
	// Class 54 — Program Limit Exceeded
	ProgramLimitExceeded PSQLError = "54000"
	StatementTooComplex  PSQLError = "54001"
	TooManyColumns       PSQLError = "54011"
	TooManyArguments     PSQLError = "54023"
	// Class 55 — Object Not In Prerequisite State
	ObjectNotInPrerequisiteState PSQLError = "55000"
	ObjectInUse                  PSQLError = "55006"
	CantChangeRuntimeParam       PSQLError = "55P02"
	LockNotAvailable             PSQLError = "55P03"
	// Class 57 — Operator Intervention
	OperatorIntervention PSQLError = "57000"
	QueryCanceled        PSQLError = "57014"
	AdminShutdown        PSQLError = "57P01"
	CrashShutdown        PSQLError = "57P02"
	CannotConnectNow     PSQLError = "57P03"
	DatabaseDropped      PSQLError = "57P04"
	// Class 58 — System Error (errors external to PostgreSQL itself)
	SystemError   PSQLError = "58000"
	IoError       PSQLError = "58030"
	UndefinedFile PSQLError = "58P01"
	DuplicateFile PSQLError = "58P02"
	// Class F0 — Configuration File Error
	ConfigFileError PSQLError = "F0000"
	LockFileExists  PSQLError = "F0001"
	// Class HV — Foreign Data Wrapper Error (SQL/MED)
	FdwError                             PSQLError = "HV000"
	FdwColumnNameNotFound                PSQLError = "HV005"
	FdwDynamicParameterValueNeeded       PSQLError = "HV002"
	FdwFunctionSequenceError             PSQLError = "HV010"
	FdwInconsistentDescriptorInformation PSQLError = "HV021"
	FdwInvalidAttributeValue             PSQLError = "HV024"
	FdwInvalidColumnName                 PSQLError = "HV007"
	FdwInvalidColumnNumber               PSQLError = "HV008"
	FdwInvalidDataType                   PSQLError = "HV004"
	FdwInvalidDataTypeDescriptors        PSQLError = "HV006"
	FdwInvalidDescriptorFieldIdentifier  PSQLError = "HV091"
	FdwInvalidHandle                     PSQLError = "HV00B"
	FdwInvalidOptionIndex                PSQLError = "HV00C"
	FdwInvalidOptionName                 PSQLError = "HV00D"
	FdwInvalidStringLengthOrBufferLength PSQLError = "HV090"
	FdwInvalidStringFormat               PSQLError = "HV00A"
	FdwInvalidUseOfNullPointer           PSQLError = "HV009"
	FdwTooManyHandles                    PSQLError = "HV014"
	FdwOutOfMemory                       PSQLError = "HV001"
	FdwNoSchemas                         PSQLError = "HV00P"
	FdwOptionNameNotFound                PSQLError = "HV00J"
	FdwReplyHandle                       PSQLError = "HV00K"
	FdwSchemaNotFound                    PSQLError = "HV00Q"
	FdwTableNotFound                     PSQLError = "HV00R"
	FdwUnableToCreateExecution           PSQLError = "HV00L"
	FdwUnableToCreateReply               PSQLError = "HV00M"
	FdwUnableToEstablishConnection       PSQLError = "HV00N"
	// Class P0 — PL/pgSQL Error
	PlpgsqlError   PSQLError = "P0000"
	RaiseException PSQLError = "P0001"
	NoDataFound    PSQLError = "P0002"
	TooManyRows    PSQLError = "P0003"
	// Class XX — Internal Error
	InternalError  PSQLError = "XX000"
	DataCorrupted  PSQLError = "XX001"
	IndexCorrupted PSQLError = "XX002"
)

func (e *PSQLError) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	default:
		*e = UnknownPSQLError
	case "successful_completion":
		*e = SuccessfulCompletion
	case "warning":
		*e = Warning
	case "dynamic_result_sets_returned":
		*e = DynamicResultSetsReturned
	case "implicit_zero_bit_padding":
		*e = ImplicitZeroBitPadding
	case "null_value_eliminated_in_set_function":
		*e = NullValueEliminatedInSetFunction
	case "privilege_not_granted":
		*e = PrivilegeNotGranted
	case "privilege_not_revoked":
		*e = PrivilegeNotRevoked
	case "warning_string_data_right_truncation":
		*e = WarningStringDataRightTruncation
	case "deprecated_feature":
		*e = DeprecatedFeature
	case "no_data":
		*e = NoData
	case "no_additional_dynamic_result_sets_returned":
		*e = NoAdditionalDynamicResultSetsReturned
	case "sql_statement_not_yet_complete":
		*e = SqlStatementNotYetComplete
	case "connection_exception":
		*e = ConnectionException
	case "connection_does_not_exist":
		*e = ConnectionDoesNotExist
	case "connection_failure":
		*e = ConnectionFailure
	case "sqlclient_unable_to_establish_sqlconnection":
		*e = SqlclientUnableToEstablishSqlconnection
	case "sqlserver_rejected_establishment_of_sqlconnection":
		*e = SqlserverRejectedEstablishmentOfSqlconnection
	case "transaction_resolution_unknown":
		*e = TransactionResolutionUnknown
	case "protocol_violation":
		*e = ProtocolViolation
	case "triggered_action_exception":
		*e = TriggeredActionException
	case "feature_not_supported":
		*e = FeatureNotSupported
	case "invalid_transaction_initiation":
		*e = InvalidTransactionInitiation
	case "locator_exception":
		*e = LocatorException
	case "invalid_locator_specification":
		*e = InvalidLocatorSpecification
	case "invalid_grantor":
		*e = InvalidGrantor
	case "invalid_grant_operation":
		*e = InvalidGrantOperation
	case "invalid_role_specification":
		*e = InvalidRoleSpecification
	case "diagnostics_exception":
		*e = DiagnosticsException
	case "stacked_diagnostics_accessed_without_active_handler":
		*e = StackedDiagnosticsAccessedWithoutActiveHandler
	case "case_not_found":
		*e = CaseNotFound
	case "cardinality_violation":
		*e = CardinalityViolation
	case "data_exception":
		*e = DataException
	case "array_subscript_error":
		*e = ArraySubscriptError
	case "character_not_in_repertoire":
		*e = CharacterNotInRepertoire
	case "datetime_field_overflow":
		*e = DatetimeFieldOverflow
	case "division_by_zero":
		*e = DivisionByZero
	case "error_in_assignment":
		*e = ErrorInAssignment
	case "escape_character_conflict":
		*e = EscapeCharacterConflict
	case "indicator_overflow":
		*e = IndicatorOverflow
	case "interval_field_overflow":
		*e = IntervalFieldOverflow
	case "invalid_argument_for_logarithm":
		*e = InvalidArgumentForLogarithm
	case "invalid_argument_for_ntile_function":
		*e = InvalidArgumentForNtileFunction
	case "invalid_argument_for_nth_value_function":
		*e = InvalidArgumentForNthValueFunction
	case "invalid_argument_for_power_function":
		*e = InvalidArgumentForPowerFunction
	case "invalid_argument_for_width_bucket_function":
		*e = InvalidArgumentForWidthBucketFunction
	case "invalid_character_value_for_cast":
		*e = InvalidCharacterValueForCast
	case "invalid_datetime_format":
		*e = InvalidDatetimeFormat
	case "invalid_escape_character":
		*e = InvalidEscapeCharacter
	case "invalid_escape_octet":
		*e = InvalidEscapeOctet
	case "invalid_escape_sequence":
		*e = InvalidEscapeSequence
	case "nonstandard_use_of_escape_character":
		*e = NonstandardUseOfEscapeCharacter
	case "invalid_indicator_parameter_value":
		*e = InvalidIndicatorParameterValue
	case "invalid_parameter_value":
		*e = InvalidParameterValue
	case "invalid_regular_expression":
		*e = InvalidRegularExpression
	case "invalid_row_count_in_limit_clause":
		*e = InvalidRowCountInLimitClause
	case "invalid_row_count_in_result_offset_clause":
		*e = InvalidRowCountInResultOffsetClause
	case "invalid_time_zone_displacement_value":
		*e = InvalidTimeZoneDisplacementValue
	case "invalid_use_of_escape_character":
		*e = InvalidUseOfEscapeCharacter
	case "most_specific_type_mismatch":
		*e = MostSpecificTypeMismatch
	case "null_value_not_allowed":
		*e = NullValueNotAllowed
	case "null_value_no_indicator_parameter":
		*e = NullValueNoIndicatorParameter
	case "numeric_value_out_of_range":
		*e = NumericValueOutOfRange
	case "string_data_length_mismatch":
		*e = StringDataLengthMismatch
	case "string_data_right_truncation":
		*e = StringDataRightTruncation
	case "substring_error":
		*e = SubstringError
	case "trim_error":
		*e = TrimError
	case "unterminated_c_string":
		*e = UnterminatedCString
	case "zero_length_character_string":
		*e = ZeroLengthCharacterString
	case "floating_point_exception":
		*e = FloatingPointException
	case "invalid_text_representation":
		*e = InvalidTextRepresentation
	case "invalid_binary_representation":
		*e = InvalidBinaryRepresentation
	case "bad_copy_file_format":
		*e = BadCopyFileFormat
	case "untranslatable_character":
		*e = UntranslatableCharacter
	case "not_an_xml_document":
		*e = NotAnXmlDocument
	case "invalid_xml_document":
		*e = InvalidXmlDocument
	case "invalid_xml_content":
		*e = InvalidXmlContent
	case "invalid_xml_comment":
		*e = InvalidXmlComment
	case "invalid_xml_processing_instruction":
		*e = InvalidXmlProcessingInstruction
	case "integrity_constraint_violation":
		*e = IntegrityConstraintViolation
	case "restrict_violation":
		*e = RestrictViolation
	case "not_null_violation":
		*e = NotNullViolation
	case "foreign_key_violation":
		*e = ForeignKeyViolation
	case "unique_violation":
		*e = UniqueViolation
	case "check_violation":
		*e = CheckViolation
	case "exclusion_violation":
		*e = ExclusionViolation
	case "invalid_cursor_state":
		*e = InvalidCursorState
	case "invalid_transaction_state":
		*e = InvalidTransactionState
	case "active_sql_transaction":
		*e = ActiveSqlTransaction
	case "branch_transaction_already_active":
		*e = BranchTransactionAlreadyActive
	case "held_cursor_requires_same_isolation_level":
		*e = HeldCursorRequiresSameIsolationLevel
	case "inappropriate_access_mode_for_branch_transaction":
		*e = InappropriateAccessModeForBranchTransaction
	case "inappropriate_isolation_level_for_branch_transaction":
		*e = InappropriateIsolationLevelForBranchTransaction
	case "no_active_sql_transaction_for_branch_transaction":
		*e = NoActiveSqlTransactionForBranchTransaction
	case "read_only_sql_transaction":
		*e = ReadOnlySqlTransaction
	case "schema_and_data_statement_mixing_not_supported":
		*e = SchemaAndDataStatementMixingNotSupported
	case "no_active_sql_transaction":
		*e = NoActiveSqlTransaction
	case "in_failed_sql_transaction":
		*e = InFailedSqlTransaction
	case "invalid_sql_statement_name":
		*e = InvalidSqlStatementName
	case "triggered_data_change_violation":
		*e = TriggeredDataChangeViolation
	case "invalid_authorization_specification":
		*e = InvalidAuthorizationSpecification
	case "invalid_password":
		*e = InvalidPassword
	case "dependent_privilege_descriptors_still_exist":
		*e = DependentPrivilegeDescriptorsStillExist
	case "dependent_objects_still_exist":
		*e = DependentObjectsStillExist
	case "invalid_transaction_termination":
		*e = InvalidTransactionTermination
	case "sql_routine_exception":
		*e = SqlRoutineException
	case "function_executed_no_return_statement":
		*e = FunctionExecutedNoReturnStatement
	case "modifying_sql_data_not_permitted":
		*e = ModifyingSqlDataNotPermitted
	case "prohibited_sql_statement_attempted":
		*e = ProhibitedSqlStatementAttempted
	case "reading_sql_data_not_permitted":
		*e = ReadingSqlDataNotPermitted
	case "invalid_cursor_name":
		*e = InvalidCursorName
	case "external_routine_exception":
		*e = ExternalRoutineException
	case "containing_sql_not_permitted":
		*e = ContainingSqlNotPermitted
	case "ere_modifying_sql_data_not_permitted":
		*e = EreModifyingSqlDataNotPermitted
	case "ere_prohibited_sql_statement_attempted":
		*e = EreProhibitedSqlStatementAttempted
	case "ere_reading_sql_data_not_permitted":
		*e = EreReadingSqlDataNotPermitted
	case "external_routine_invocation_exception":
		*e = ExternalRoutineInvocationException
	case "invalid_sqlstate_returned":
		*e = InvalidSqlstateReturned
	case "erie_null_value_not_allowed":
		*e = ErieNullValueNotAllowed
	case "trigger_protocol_violated":
		*e = TriggerProtocolViolated
	case "srf_protocol_violated":
		*e = SrfProtocolViolated
	case "savepoint_exception":
		*e = SavepointException
	case "invalid_savepoint_specification":
		*e = InvalidSavepointSpecification
	case "invalid_catalog_name":
		*e = InvalidCatalogName
	case "invalid_schema_name":
		*e = InvalidSchemaName
	case "transaction_rollback":
		*e = TransactionRollback
	case "transaction_integrity_constraint_violation":
		*e = TransactionIntegrityConstraintViolation
	case "serialization_failure":
		*e = SerializationFailure
	case "statement_completion_unknown":
		*e = StatementCompletionUnknown
	case "deadlock_detected":
		*e = DeadlockDetected
	case "syntax_error_or_access_rule_violation":
		*e = SyntaxErrorOrAccessRuleViolation
	case "syntax_error":
		*e = SyntaxError
	case "insufficient_privilege":
		*e = InsufficientPrivilege
	case "cannot_coerce":
		*e = CannotCoerce
	case "grouping_error":
		*e = GroupingError
	case "windowing_error":
		*e = WindowingError
	case "invalid_recursion":
		*e = InvalidRecursion
	case "invalid_foreign_key":
		*e = InvalidForeignKey
	case "invalid_name":
		*e = InvalidName
	case "name_too_long":
		*e = NameTooLong
	case "reserved_name":
		*e = ReservedName
	case "datatype_mismatch":
		*e = DatatypeMismatch
	case "indeterminate_datatype":
		*e = IndeterminateDatatype
	case "collation_mismatch":
		*e = CollationMismatch
	case "indeterminate_collation":
		*e = IndeterminateCollation
	case "wrong_object_type":
		*e = WrongObjectType
	case "undefined_column":
		*e = UndefinedColumn
	case "undefined_function":
		*e = UndefinedFunction
	case "undefined_table":
		*e = UndefinedTable
	case "undefined_parameter":
		*e = UndefinedParameter
	case "undefined_object":
		*e = UndefinedObject
	case "duplicate_column":
		*e = DuplicateColumn
	case "duplicate_cursor":
		*e = DuplicateCursor
	case "duplicate_database":
		*e = DuplicateDatabase
	case "duplicate_function":
		*e = DuplicateFunction
	case "duplicate_prepared_statement":
		*e = DuplicatePreparedStatement
	case "duplicate_schema":
		*e = DuplicateSchema
	case "duplicate_table":
		*e = DuplicateTable
	case "duplicate_alias":
		*e = DuplicateAlias
	case "duplicate_object":
		*e = DuplicateObject
	case "ambiguous_column":
		*e = AmbiguousColumn
	case "ambiguous_function":
		*e = AmbiguousFunction
	case "ambiguous_parameter":
		*e = AmbiguousParameter
	case "ambiguous_alias":
		*e = AmbiguousAlias
	case "invalid_column_reference":
		*e = InvalidColumnReference
	case "invalid_column_definition":
		*e = InvalidColumnDefinition
	case "invalid_cursor_definition":
		*e = InvalidCursorDefinition
	case "invalid_database_definition":
		*e = InvalidDatabaseDefinition
	case "invalid_function_definition":
		*e = InvalidFunctionDefinition
	case "invalid_prepared_statement_definition":
		*e = InvalidPreparedStatementDefinition
	case "invalid_schema_definition":
		*e = InvalidSchemaDefinition
	case "invalid_table_definition":
		*e = InvalidTableDefinition
	case "invalid_object_definition":
		*e = InvalidObjectDefinition
	case "with_check_option_violation":
		*e = WithCheckOptionViolation
	case "insufficient_resources":
		*e = InsufficientResources
	case "disk_full":
		*e = DiskFull
	case "out_of_memory":
		*e = OutOfMemory
	case "too_many_connections":
		*e = TooManyConnections
	case "configuration_limit_exceeded":
		*e = ConfigurationLimitExceeded
	case "program_limit_exceeded":
		*e = ProgramLimitExceeded
	case "statement_too_complex":
		*e = StatementTooComplex
	case "too_many_columns":
		*e = TooManyColumns
	case "too_many_arguments":
		*e = TooManyArguments
	case "object_not_in_prerequisite_state":
		*e = ObjectNotInPrerequisiteState
	case "object_in_use":
		*e = ObjectInUse
	case "cant_change_runtime_param":
		*e = CantChangeRuntimeParam
	case "lock_not_available":
		*e = LockNotAvailable
	case "operator_intervention":
		*e = OperatorIntervention
	case "query_canceled":
		*e = QueryCanceled
	case "admin_shutdown":
		*e = AdminShutdown
	case "crash_shutdown":
		*e = CrashShutdown
	case "cannot_connect_now":
		*e = CannotConnectNow
	case "database_dropped":
		*e = DatabaseDropped
	case "system_error":
		*e = SystemError
	case "io_error":
		*e = IoError
	case "undefined_file":
		*e = UndefinedFile
	case "duplicate_file":
		*e = DuplicateFile
	case "config_file_error":
		*e = ConfigFileError
	case "lock_file_exists":
		*e = LockFileExists
	case "fdw_error":
		*e = FdwError
	case "fdw_column_name_not_found":
		*e = FdwColumnNameNotFound
	case "fdw_dynamic_parameter_value_needed":
		*e = FdwDynamicParameterValueNeeded
	case "fdw_function_sequence_error":
		*e = FdwFunctionSequenceError
	case "fdw_inconsistent_descriptor_information":
		*e = FdwInconsistentDescriptorInformation
	case "fdw_invalid_attribute_value":
		*e = FdwInvalidAttributeValue
	case "fdw_invalid_column_name":
		*e = FdwInvalidColumnName
	case "fdw_invalid_column_number":
		*e = FdwInvalidColumnNumber
	case "fdw_invalid_data_type":
		*e = FdwInvalidDataType
	case "fdw_invalid_data_type_descriptors":
		*e = FdwInvalidDataTypeDescriptors
	case "fdw_invalid_descriptor_field_identifier":
		*e = FdwInvalidDescriptorFieldIdentifier
	case "fdw_invalid_handle":
		*e = FdwInvalidHandle
	case "fdw_invalid_option_index":
		*e = FdwInvalidOptionIndex
	case "fdw_invalid_option_name":
		*e = FdwInvalidOptionName
	case "fdw_invalid_string_length_or_buffer_length":
		*e = FdwInvalidStringLengthOrBufferLength
	case "fdw_invalid_string_format":
		*e = FdwInvalidStringFormat
	case "fdw_invalid_use_of_null_pointer":
		*e = FdwInvalidUseOfNullPointer
	case "fdw_too_many_handles":
		*e = FdwTooManyHandles
	case "fdw_out_of_memory":
		*e = FdwOutOfMemory
	case "fdw_no_schemas":
		*e = FdwNoSchemas
	case "fdw_option_name_not_found":
		*e = FdwOptionNameNotFound
	case "fdw_reply_handle":
		*e = FdwReplyHandle
	case "fdw_schema_not_found":
		*e = FdwSchemaNotFound
	case "fdw_table_not_found":
		*e = FdwTableNotFound
	case "fdw_unable_to_create_execution":
		*e = FdwUnableToCreateExecution
	case "fdw_unable_to_create_reply":
		*e = FdwUnableToCreateReply
	case "fdw_unable_to_establish_connection":
		*e = FdwUnableToEstablishConnection
	case "plpgsql_error":
		*e = PlpgsqlError
	case "raise_exception":
		*e = RaiseException
	case "no_data_found":
		*e = NoDataFound
	case "too_many_rows":
		*e = TooManyRows
	case "internal_error":
		*e = InternalError
	case "data_corrupted":
		*e = DataCorrupted
	case "index_corrupted":
		*e = IndexCorrupted
	}

	return nil
}

func (e PSQLError) MarshalJSON() ([]byte, error) {
	var s string
	switch e {
	default:
		s = "unknown_psql_error"
	case SuccessfulCompletion:
		s = "successful_completion"
	case Warning:
		s = "warning"
	case DynamicResultSetsReturned:
		s = "dynamic_result_sets_returned"
	case ImplicitZeroBitPadding:
		s = "implicit_zero_bit_padding"
	case NullValueEliminatedInSetFunction:
		s = "null_value_eliminated_in_set_function"
	case PrivilegeNotGranted:
		s = "privilege_not_granted"
	case PrivilegeNotRevoked:
		s = "privilege_not_revoked"
	case WarningStringDataRightTruncation:
		s = "warning_string_data_right_truncation"
	case DeprecatedFeature:
		s = "deprecated_feature"
	case NoData:
		s = "no_data"
	case NoAdditionalDynamicResultSetsReturned:
		s = "no_additional_dynamic_result_sets_returned"
	case SqlStatementNotYetComplete:
		s = "sql_statement_not_yet_complete"
	case ConnectionException:
		s = "connection_exception"
	case ConnectionDoesNotExist:
		s = "connection_does_not_exist"
	case ConnectionFailure:
		s = "connection_failure"
	case SqlclientUnableToEstablishSqlconnection:
		s = "sqlclient_unable_to_establish_sqlconnection"
	case SqlserverRejectedEstablishmentOfSqlconnection:
		s = "sqlserver_rejected_establishment_of_sqlconnection"
	case TransactionResolutionUnknown:
		s = "transaction_resolution_unknown"
	case ProtocolViolation:
		s = "protocol_violation"
	case TriggeredActionException:
		s = "triggered_action_exception"
	case FeatureNotSupported:
		s = "feature_not_supported"
	case InvalidTransactionInitiation:
		s = "invalid_transaction_initiation"
	case LocatorException:
		s = "locator_exception"
	case InvalidLocatorSpecification:
		s = "invalid_locator_specification"
	case InvalidGrantor:
		s = "invalid_grantor"
	case InvalidGrantOperation:
		s = "invalid_grant_operation"
	case InvalidRoleSpecification:
		s = "invalid_role_specification"
	case DiagnosticsException:
		s = "diagnostics_exception"
	case StackedDiagnosticsAccessedWithoutActiveHandler:
		s = "stacked_diagnostics_accessed_without_active_handler"
	case CaseNotFound:
		s = "case_not_found"
	case CardinalityViolation:
		s = "cardinality_violation"
	case DataException:
		s = "data_exception"
	case ArraySubscriptError:
		s = "array_subscript_error"
	case CharacterNotInRepertoire:
		s = "character_not_in_repertoire"
	case DatetimeFieldOverflow:
		s = "datetime_field_overflow"
	case DivisionByZero:
		s = "division_by_zero"
	case ErrorInAssignment:
		s = "error_in_assignment"
	case EscapeCharacterConflict:
		s = "escape_character_conflict"
	case IndicatorOverflow:
		s = "indicator_overflow"
	case IntervalFieldOverflow:
		s = "interval_field_overflow"
	case InvalidArgumentForLogarithm:
		s = "invalid_argument_for_logarithm"
	case InvalidArgumentForNtileFunction:
		s = "invalid_argument_for_ntile_function"
	case InvalidArgumentForNthValueFunction:
		s = "invalid_argument_for_nth_value_function"
	case InvalidArgumentForPowerFunction:
		s = "invalid_argument_for_power_function"
	case InvalidArgumentForWidthBucketFunction:
		s = "invalid_argument_for_width_bucket_function"
	case InvalidCharacterValueForCast:
		s = "invalid_character_value_for_cast"
	case InvalidDatetimeFormat:
		s = "invalid_datetime_format"
	case InvalidEscapeCharacter:
		s = "invalid_escape_character"
	case InvalidEscapeOctet:
		s = "invalid_escape_octet"
	case InvalidEscapeSequence:
		s = "invalid_escape_sequence"
	case NonstandardUseOfEscapeCharacter:
		s = "nonstandard_use_of_escape_character"
	case InvalidIndicatorParameterValue:
		s = "invalid_indicator_parameter_value"
	case InvalidParameterValue:
		s = "invalid_parameter_value"
	case InvalidRegularExpression:
		s = "invalid_regular_expression"
	case InvalidRowCountInLimitClause:
		s = "invalid_row_count_in_limit_clause"
	case InvalidRowCountInResultOffsetClause:
		s = "invalid_row_count_in_result_offset_clause"
	case InvalidTimeZoneDisplacementValue:
		s = "invalid_time_zone_displacement_value"
	case InvalidUseOfEscapeCharacter:
		s = "invalid_use_of_escape_character"
	case MostSpecificTypeMismatch:
		s = "most_specific_type_mismatch"
	case NullValueNotAllowed:
		s = "null_value_not_allowed"
	case NullValueNoIndicatorParameter:
		s = "null_value_no_indicator_parameter"
	case NumericValueOutOfRange:
		s = "numeric_value_out_of_range"
	case StringDataLengthMismatch:
		s = "string_data_length_mismatch"
	case StringDataRightTruncation:
		s = "string_data_right_truncation"
	case SubstringError:
		s = "substring_error"
	case TrimError:
		s = "trim_error"
	case UnterminatedCString:
		s = "unterminated_c_string"
	case ZeroLengthCharacterString:
		s = "zero_length_character_string"
	case FloatingPointException:
		s = "floating_point_exception"
	case InvalidTextRepresentation:
		s = "invalid_text_representation"
	case InvalidBinaryRepresentation:
		s = "invalid_binary_representation"
	case BadCopyFileFormat:
		s = "bad_copy_file_format"
	case UntranslatableCharacter:
		s = "untranslatable_character"
	case NotAnXmlDocument:
		s = "not_an_xml_document"
	case InvalidXmlDocument:
		s = "invalid_xml_document"
	case InvalidXmlContent:
		s = "invalid_xml_content"
	case InvalidXmlComment:
		s = "invalid_xml_comment"
	case InvalidXmlProcessingInstruction:
		s = "invalid_xml_processing_instruction"
	case IntegrityConstraintViolation:
		s = "integrity_constraint_violation"
	case RestrictViolation:
		s = "restrict_violation"
	case NotNullViolation:
		s = "not_null_violation"
	case ForeignKeyViolation:
		s = "foreign_key_violation"
	case UniqueViolation:
		s = "unique_violation"
	case CheckViolation:
		s = "check_violation"
	case ExclusionViolation:
		s = "exclusion_violation"
	case InvalidCursorState:
		s = "invalid_cursor_state"
	case InvalidTransactionState:
		s = "invalid_transaction_state"
	case ActiveSqlTransaction:
		s = "active_sql_transaction"
	case BranchTransactionAlreadyActive:
		s = "branch_transaction_already_active"
	case HeldCursorRequiresSameIsolationLevel:
		s = "held_cursor_requires_same_isolation_level"
	case InappropriateAccessModeForBranchTransaction:
		s = "inappropriate_access_mode_for_branch_transaction"
	case InappropriateIsolationLevelForBranchTransaction:
		s = "inappropriate_isolation_level_for_branch_transaction"
	case NoActiveSqlTransactionForBranchTransaction:
		s = "no_active_sql_transaction_for_branch_transaction"
	case ReadOnlySqlTransaction:
		s = "read_only_sql_transaction"
	case SchemaAndDataStatementMixingNotSupported:
		s = "schema_and_data_statement_mixing_not_supported"
	case NoActiveSqlTransaction:
		s = "no_active_sql_transaction"
	case InFailedSqlTransaction:
		s = "in_failed_sql_transaction"
	case InvalidSqlStatementName:
		s = "invalid_sql_statement_name"
	case TriggeredDataChangeViolation:
		s = "triggered_data_change_violation"
	case InvalidAuthorizationSpecification:
		s = "invalid_authorization_specification"
	case InvalidPassword:
		s = "invalid_password"
	case DependentPrivilegeDescriptorsStillExist:
		s = "dependent_privilege_descriptors_still_exist"
	case DependentObjectsStillExist:
		s = "dependent_objects_still_exist"
	case InvalidTransactionTermination:
		s = "invalid_transaction_termination"
	case SqlRoutineException:
		s = "sql_routine_exception"
	case FunctionExecutedNoReturnStatement:
		s = "function_executed_no_return_statement"
	case ModifyingSqlDataNotPermitted:
		s = "modifying_sql_data_not_permitted"
	case ProhibitedSqlStatementAttempted:
		s = "prohibited_sql_statement_attempted"
	case ReadingSqlDataNotPermitted:
		s = "reading_sql_data_not_permitted"
	case InvalidCursorName:
		s = "invalid_cursor_name"
	case ExternalRoutineException:
		s = "external_routine_exception"
	case ContainingSqlNotPermitted:
		s = "containing_sql_not_permitted"
	case EreModifyingSqlDataNotPermitted:
		s = "ere_modifying_sql_data_not_permitted"
	case EreProhibitedSqlStatementAttempted:
		s = "ere_prohibited_sql_statement_attempted"
	case EreReadingSqlDataNotPermitted:
		s = "ere_reading_sql_data_not_permitted"
	case ExternalRoutineInvocationException:
		s = "external_routine_invocation_exception"
	case InvalidSqlstateReturned:
		s = "invalid_sqlstate_returned"
	case ErieNullValueNotAllowed:
		s = "erie_null_value_not_allowed"
	case TriggerProtocolViolated:
		s = "trigger_protocol_violated"
	case SrfProtocolViolated:
		s = "srf_protocol_violated"
	case SavepointException:
		s = "savepoint_exception"
	case InvalidSavepointSpecification:
		s = "invalid_savepoint_specification"
	case InvalidCatalogName:
		s = "invalid_catalog_name"
	case InvalidSchemaName:
		s = "invalid_schema_name"
	case TransactionRollback:
		s = "transaction_rollback"
	case TransactionIntegrityConstraintViolation:
		s = "transaction_integrity_constraint_violation"
	case SerializationFailure:
		s = "serialization_failure"
	case StatementCompletionUnknown:
		s = "statement_completion_unknown"
	case DeadlockDetected:
		s = "deadlock_detected"
	case SyntaxErrorOrAccessRuleViolation:
		s = "syntax_error_or_access_rule_violation"
	case SyntaxError:
		s = "syntax_error"
	case InsufficientPrivilege:
		s = "insufficient_privilege"
	case CannotCoerce:
		s = "cannot_coerce"
	case GroupingError:
		s = "grouping_error"
	case WindowingError:
		s = "windowing_error"
	case InvalidRecursion:
		s = "invalid_recursion"
	case InvalidForeignKey:
		s = "invalid_foreign_key"
	case InvalidName:
		s = "invalid_name"
	case NameTooLong:
		s = "name_too_long"
	case ReservedName:
		s = "reserved_name"
	case DatatypeMismatch:
		s = "datatype_mismatch"
	case IndeterminateDatatype:
		s = "indeterminate_datatype"
	case CollationMismatch:
		s = "collation_mismatch"
	case IndeterminateCollation:
		s = "indeterminate_collation"
	case WrongObjectType:
		s = "wrong_object_type"
	case UndefinedColumn:
		s = "undefined_column"
	case UndefinedFunction:
		s = "undefined_function"
	case UndefinedTable:
		s = "undefined_table"
	case UndefinedParameter:
		s = "undefined_parameter"
	case UndefinedObject:
		s = "undefined_object"
	case DuplicateColumn:
		s = "duplicate_column"
	case DuplicateCursor:
		s = "duplicate_cursor"
	case DuplicateDatabase:
		s = "duplicate_database"
	case DuplicateFunction:
		s = "duplicate_function"
	case DuplicatePreparedStatement:
		s = "duplicate_prepared_statement"
	case DuplicateSchema:
		s = "duplicate_schema"
	case DuplicateTable:
		s = "duplicate_table"
	case DuplicateAlias:
		s = "duplicate_alias"
	case DuplicateObject:
		s = "duplicate_object"
	case AmbiguousColumn:
		s = "ambiguous_column"
	case AmbiguousFunction:
		s = "ambiguous_function"
	case AmbiguousParameter:
		s = "ambiguous_parameter"
	case AmbiguousAlias:
		s = "ambiguous_alias"
	case InvalidColumnReference:
		s = "invalid_column_reference"
	case InvalidColumnDefinition:
		s = "invalid_column_definition"
	case InvalidCursorDefinition:
		s = "invalid_cursor_definition"
	case InvalidDatabaseDefinition:
		s = "invalid_database_definition"
	case InvalidFunctionDefinition:
		s = "invalid_function_definition"
	case InvalidPreparedStatementDefinition:
		s = "invalid_prepared_statement_definition"
	case InvalidSchemaDefinition:
		s = "invalid_schema_definition"
	case InvalidTableDefinition:
		s = "invalid_table_definition"
	case InvalidObjectDefinition:
		s = "invalid_object_definition"
	case WithCheckOptionViolation:
		s = "with_check_option_violation"
	case InsufficientResources:
		s = "insufficient_resources"
	case DiskFull:
		s = "disk_full"
	case OutOfMemory:
		s = "out_of_memory"
	case TooManyConnections:
		s = "too_many_connections"
	case ConfigurationLimitExceeded:
		s = "configuration_limit_exceeded"
	case ProgramLimitExceeded:
		s = "program_limit_exceeded"
	case StatementTooComplex:
		s = "statement_too_complex"
	case TooManyColumns:
		s = "too_many_columns"
	case TooManyArguments:
		s = "too_many_arguments"
	case ObjectNotInPrerequisiteState:
		s = "object_not_in_prerequisite_state"
	case ObjectInUse:
		s = "object_in_use"
	case CantChangeRuntimeParam:
		s = "cant_change_runtime_param"
	case LockNotAvailable:
		s = "lock_not_available"
	case OperatorIntervention:
		s = "operator_intervention"
	case QueryCanceled:
		s = "query_canceled"
	case AdminShutdown:
		s = "admin_shutdown"
	case CrashShutdown:
		s = "crash_shutdown"
	case CannotConnectNow:
		s = "cannot_connect_now"
	case DatabaseDropped:
		s = "database_dropped"
	case SystemError:
		s = "system_error"
	case IoError:
		s = "io_error"
	case UndefinedFile:
		s = "undefined_file"
	case DuplicateFile:
		s = "duplicate_file"
	case ConfigFileError:
		s = "config_file_error"
	case LockFileExists:
		s = "lock_file_exists"
	case FdwError:
		s = "fdw_error"
	case FdwColumnNameNotFound:
		s = "fdw_column_name_not_found"
	case FdwDynamicParameterValueNeeded:
		s = "fdw_dynamic_parameter_value_needed"
	case FdwFunctionSequenceError:
		s = "fdw_function_sequence_error"
	case FdwInconsistentDescriptorInformation:
		s = "fdw_inconsistent_descriptor_information"
	case FdwInvalidAttributeValue:
		s = "fdw_invalid_attribute_value"
	case FdwInvalidColumnName:
		s = "fdw_invalid_column_name"
	case FdwInvalidColumnNumber:
		s = "fdw_invalid_column_number"
	case FdwInvalidDataType:
		s = "fdw_invalid_data_type"
	case FdwInvalidDataTypeDescriptors:
		s = "fdw_invalid_data_type_descriptors"
	case FdwInvalidDescriptorFieldIdentifier:
		s = "fdw_invalid_descriptor_field_identifier"
	case FdwInvalidHandle:
		s = "fdw_invalid_handle"
	case FdwInvalidOptionIndex:
		s = "fdw_invalid_option_index"
	case FdwInvalidOptionName:
		s = "fdw_invalid_option_name"
	case FdwInvalidStringLengthOrBufferLength:
		s = "fdw_invalid_string_length_or_buffer_length"
	case FdwInvalidStringFormat:
		s = "fdw_invalid_string_format"
	case FdwInvalidUseOfNullPointer:
		s = "fdw_invalid_use_of_null_pointer"
	case FdwTooManyHandles:
		s = "fdw_too_many_handles"
	case FdwOutOfMemory:
		s = "fdw_out_of_memory"
	case FdwNoSchemas:
		s = "fdw_no_schemas"
	case FdwOptionNameNotFound:
		s = "fdw_option_name_not_found"
	case FdwReplyHandle:
		s = "fdw_reply_handle"
	case FdwSchemaNotFound:
		s = "fdw_schema_not_found"
	case FdwTableNotFound:
		s = "fdw_table_not_found"
	case FdwUnableToCreateExecution:
		s = "fdw_unable_to_create_execution"
	case FdwUnableToCreateReply:
		s = "fdw_unable_to_create_reply"
	case FdwUnableToEstablishConnection:
		s = "fdw_unable_to_establish_connection"
	case PlpgsqlError:
		s = "plpgsql_error"
	case RaiseException:
		s = "raise_exception"
	case NoDataFound:
		s = "no_data_found"
	case TooManyRows:
		s = "too_many_rows"
	case InternalError:
		s = "internal_error"
	case DataCorrupted:
		s = "data_corrupted"
	case IndexCorrupted:
		s = "index_corrupted"
	}

	return json.Marshal(s)
}
