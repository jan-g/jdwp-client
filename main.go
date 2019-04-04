package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/jan-g/jdwp-client/client"
)

var (
	level   = flag.String("log", "debug", "log level")
	net     = flag.String("net", "tcp", "network type")
	address = flag.String("addr", "localhost:59999", "address to connect to")

	cls        = flag.String("class", "Lorg/ioctl/debug/app/WebServer$Handler;", "class to break on")
	methodName = flag.String("method", "handle", "method to break on")
	methodSig  = flag.String("signature", "(Lcom/sun/net/httpserver/HttpExchange;)V", "method signature to break on")
	line       = flag.Int("line", 23, "line number to break on")
	variable   = flag.String("variable", "response", "Variable to inspect")
)

func main() {
	flag.Parse()
	log, err := logrus.ParseLevel(*level)
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(log)
	c, err := client.Dial(*net, *address)
	if err != nil {
		panic(err)
	}
	r, err := c.Call(client.VirtualMachine, client.VirtualMachineVersion, []byte{})
	var v client.VersionReply
	client.Parse(r.Data, &v)
	fmt.Printf("Vesion: %+v\n", v)

	_, _ = referenceType(c, "Ljava/lang/String;")
	myClasses, _ := referenceType(c, *cls)
	myClass := myClasses[0]

	sig, err := myClass.Signature(c)
	fmt.Println("signature of class is", err, sig)

	// Get methods
	ms, err := myClass.Methods(c)
	fmt.Println("class methods are", err, ms)

	m := ms[0]
	for _, mm := range ms {
		if mm.Name == *methodName && mm.Signature == *methodSig {
			m = mm
			break
		}
	}

	// Get line map
	lines, err := m.MethodId.LineTable(c)
	fmt.Println("line table:", err, lines)

	l := lines.LineEntries[0]
	for _, ll := range lines.LineEntries {
		if ll.LineNumber == *line {
			l = ll
		}
	}

	// Construct the location
	location := client.NewLocation(m.MethodId, l.LineCodeIndex)

	r, err = c.Call(client.EventRequest, client.Set,
		client.NewEventRequestSet(client.EventKindBreakpoint, client.SuspendPolicyEventThread).
			WithMod(client.ModKindLocation).WithLocation(location).
			WithMod(client.ModKindCount).WithInt(1).
			Marshal())
	var bp client.EventRequestSetReply
	err = client.Parse(r.Data, &bp)
	fmt.Printf("breakpoint response received: %v, %+v -> %+v\n", err, *r, bp)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			select {
			case e, ok := <-c.Events():
				if !ok {
					fmt.Println("events end, exiting")
					wg.Done()
					return
				}
				logrus.Debugf("event received: %+v\n", *e)
				var comp client.Composite
				err := client.Parse(e.Data, &comp)

				bp := comp.Events[0].(*client.EventBreakpoint)
				fmt.Printf("composite received: %v, %+v %+v\n", err, comp, bp)

				frames, err := bp.Thread.Frames(c, 0, 3)
				fmt.Printf("frames %v %+v", err, frames)

				for _, f := range frames {
					sig, err := f.Location.ClassId.Signature(c)
					fmt.Println("Frame: ", f.FrameId, sig, lookupMethod(c, f.Location.ClassId, f.Location.MethodId), err)
					fmt.Println(" Variables in frame:", lookupVars(c, f.Location.ClassId, f.Location.MethodId))
				}
				fmt.Println()

				// Work out the variable to get
				vars := lookupVars(c, frames[0].Location.ClassId, frames[0].Location.MethodId)

				vds := []client.VariableDef{}
				for _, vdef := range vars {
					logrus.Debugf("Variable is at: %+v\n", vdef)
					vds = append(vds, vdef)
				}
				if vs, err := frames[0].GetValues(c, vds...); err == nil {
					for vname, valRef := range vs {
						// Get the value
						logrus.Debugf("Variable %v is %v %+v\n", vname, err, valRef)
						// Recover the referent
						value, err := vs[vname].RecoverValue(c)
						if err != nil {
							logrus.WithError(err).Error("problem recovering value")
						} else {
							fmt.Printf("*** %s = %+v\n", vname, value)
							switch o := value.(type) {
							case *client.Object:
								fmt.Printf("    class = %+v\n", *o.Class)
							}
						}
					}
				} else {
					logrus.Error("Problem getting variables: ", err)
				}

				r, err := c.Call(client.Thread, client.ThreadResume,
					client.Seq().ThreadId(bp.Thread).Marshal())
				logrus.Debugf("response received to Resume: %v %+v\n", err, *r)

			}
		}
	}()

	foo := prompt("Hit return when done: ")
	fmt.Println(foo)

	r, err = c.Call(client.EventRequest, client.Clear,
		client.Seq().
			Octet(uint8(client.EventKindBreakpoint)).
			Int(bp.RequestId).
			Marshal())
	fmt.Printf("response received to Clear: %+v\n", *r)

	r, err = c.Call(client.EventRequest, client.ClearAllBreakPoints, []byte{})
	fmt.Printf("response received to ClearAllBreakpoints: %+v\n", *r)

	r, err = c.Call(client.VirtualMachine, client.VirtualMachineDispose, []byte{})
	fmt.Printf("response received to Dispose: %+v\n", *r)

	c.Close()
	wg.Wait()
}

func referenceType(c client.Client, class string) ([]client.ClassId, error) {
	r, err := c.Call(client.VirtualMachine, client.VirtualMachineClassesBySignature, client.Seq().String(class).Marshal())
	if err != nil {
		return nil, err
	}
	//logrus.Debugf("response received to ClassesBySignature: %+v\n", *r)
	var cws client.ClassesBySignatureReply
	if err := client.Parse(r.Data, &cws); err == nil {
		//logrus.Debugf("response received to ClassesBySignature: %+v\n", cws)
		rts := []client.ClassId{}
		for _, cd := range cws.ClassDetails {
			rts = append(rts, cd.ClassId)
		}
		return rts, nil
	} else {
		logrus.Errorf("response received to ClassesBySignature unmarshaling:", err)
		return nil, err
	}
}

var stdin = bufio.NewScanner(os.Stdin)

func prompt(p string) string {
	fmt.Print(p)
	stdin.Scan()
	return stdin.Text()
}

type md struct {
	*client.MethodDef
	vs map[string]client.VariableDef
}

var methodCache = map[client.ClassId]map[client.MethodId]md{}

func lookupMethod(c client.Client, classId client.ClassId, methodId client.MethodId) md {
	if mmap, ok := methodCache[classId]; ok {
		return mmap[methodId]
	} else {
		mmap = map[client.MethodId]md{}
		methodCache[classId] = mmap
		ms, err := classId.Methods(c)
		if err != nil {
			logrus.WithError(err).Error("trouble looking up class definition")
			return md{}
		}
		for _, m := range ms {
			mmap[m.MethodId] = md{MethodDef: &m}
		}
		return mmap[methodId]
	}
}

func lookupVars(c client.Client, classId client.ClassId, methodId client.MethodId) map[string]client.VariableDef {
	m := lookupMethod(c, classId, methodId)
	if m.vs != nil {
		return m.vs
	}
	vtr, err := methodId.VariableTable(c)
	if err != nil {
		logrus.WithError(err).Error("trouble looking up variable table definition")
		return nil
	}
	mm := map[string]client.VariableDef{}
	for _, v := range vtr.Variables {
		mm[v.Name] = v
	}
	methodCache[classId][methodId] = md{MethodDef: m.MethodDef, vs: mm}
	return mm
}
