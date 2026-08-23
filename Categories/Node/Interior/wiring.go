package interior

func Wire(row int32, buildFrame func(tick uint32), sceneRoot string) (*InteriorStream, *Mailbox, *Emitter) {
	stream := NewInteriorStream(buildFrame, row, SlotsPerNode)
	stream.SetValueWriter(NewValueWriter(sceneRoot, int(row)))
	stream.SetSceneRoot(sceneRoot)
	mailbox := NewMailbox(row)
	return stream, mailbox, NewEmitter(mailbox, row)
}
