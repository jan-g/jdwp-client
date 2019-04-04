package client

const (
	VirtualMachine                      = CommandSet(1)
	VirtualMachineVersion               = Command(1)
	VirtualMachineClassesBySignature    = Command(2)
	VirtualMachineAllClasses            = Command(3)
	VirtualMachineAllThreads            = Command(4)
	VirtualMachineTopLevelThreadGroups  = Command(5)
	VirtualMachineDispose               = Command(6)
	VirtualMachineIDSizes               = Command(7)
	VirtualMachineSuspend               = Command(8)
	VirtualMachineResume                = Command(9)
	VirtualMachineExit                  = Command(10)
	VirtualMachineCreateString          = Command(11)
	VirtualMachineCapabilities          = Command(12)
	VirtualMachineClassPaths            = Command(13)
	VirtualMachineDisposeObjects        = Command(14)
	VirtualMachineHoldEvents            = Command(15)
	VirtualMachineReleaseEvents         = Command(16)
	VirtualMachineCapabilitiesNew       = Command(17)
	VirtualMachineRedefineClasses       = Command(18)
	VirtualMachineSetDefaultStratum     = Command(19)
	VirtualMachineAllClassesWithGeneric = Command(20)
	VirtualMachineInstanceCounts        = Command(21)
)

type VersionReply struct {
	Description string `jdwp:"Text information on the VM version"`
	JdwpMajor   int    `jdwp:"Major JDWP Version number"`
	JdwpMinor   int    `jdwp:"Minor JDWP Version number"`
	VmVersion   string `jdwp:"Target VM JRE version, as in the java.version property"`
	VmName      string `jdwp:"Target VM name, as in the java.vm.name property"`
}

type ClassesBySignatureReply struct {
	Classes      int            // Number of reference types that follow.
	ClassDetails []ClassDetails `jdwp:"counter:Classes"`
}

type ReferenceTypeId uint64

type ClassDetails struct {
	RefTypeTag uint8
	ClassId    ClassId // Kind of following reference type.
	Status     int     //	The current class status.
}
