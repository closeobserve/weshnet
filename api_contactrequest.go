package weshnet

import (
	"context"

	"github.com/libp2p/go-libp2p/core/crypto"

	"berty.tech/weshnet/pkg/errcode"
	"berty.tech/weshnet/pkg/protocoltypes"
	"berty.tech/weshnet/pkg/tyber"
)

// ContactRequestReference retrieves the necessary information to create a contact link
func (s *service) ContactRequestReference(ctx context.Context, _ *protocoltypes.ContactRequestReference_Request) (*protocoltypes.ContactRequestReference_Reply, error) {
	accountGroup := s.getAccountGroup()
	if accountGroup == nil {
		return nil, errcode.ErrGroupMissing
	}

	enabled, shareableContact := accountGroup.MetadataStore().GetIncomingContactRequestsStatus()
	rdvSeed := []byte(nil)

	if shareableContact != nil {
		rdvSeed = shareableContact.PublicRendezvousSeed
	}

	return &protocoltypes.ContactRequestReference_Reply{
		PublicRendezvousSeed: rdvSeed,
		Enabled:              enabled,
	}, nil
}

// ContactRequestDisable disables incoming contact requests
func (s *service) ContactRequestDisable(ctx context.Context, _ *protocoltypes.ContactRequestDisable_Request) (_ *protocoltypes.ContactRequestDisable_Reply, err error) {
	ctx, _, endSection := tyber.Section(ctx, s.logger, "Disabling contact requests")
	defer func() { endSection(err, "") }()

	accountGroup := s.getAccountGroup()
	if accountGroup == nil {
		return nil, errcode.ErrGroupMissing
	}

	if _, err := accountGroup.MetadataStore().ContactRequestDisable(ctx); err != nil {
		return nil, errcode.ErrOrbitDBAppend.Wrap(err)
	}

	return &protocoltypes.ContactRequestDisable_Reply{}, nil
}

// ContactRequestEnable enables incoming contact requests
func (s *service) ContactRequestEnable(ctx context.Context, _ *protocoltypes.ContactRequestEnable_Request) (_ *protocoltypes.ContactRequestEnable_Reply, err error) {
	ctx, _, endSection := tyber.Section(ctx, s.logger, "Enabling contact requests")
	defer func() { endSection(err, "") }()

	accountGroup := s.getAccountGroup()
	if accountGroup == nil {
		return nil, errcode.ErrGroupMissing
	}

	if _, err := accountGroup.MetadataStore().ContactRequestEnable(ctx); err != nil {
		return nil, errcode.ErrOrbitDBAppend.Wrap(err)
	}

	_, shareableContact := accountGroup.MetadataStore().GetIncomingContactRequestsStatus()
	rdvSeed := []byte(nil)

	if shareableContact != nil {
		rdvSeed = shareableContact.PublicRendezvousSeed
	}

	return &protocoltypes.ContactRequestEnable_Reply{
		PublicRendezvousSeed: rdvSeed,
	}, nil
}

// ContactRequestResetReference generates a new contact request reference
func (s *service) ContactRequestResetReference(ctx context.Context, _ *protocoltypes.ContactRequestResetReference_Request) (_ *protocoltypes.ContactRequestResetReference_Reply, err error) {
	ctx, _, endSection := tyber.Section(ctx, s.logger, "Resetting contact requests reference")
	defer func() { endSection(err, "") }()

	accountGroup := s.getAccountGroup()
	if accountGroup == nil {
		return nil, errcode.ErrGroupMissing
	}

	if _, err := accountGroup.MetadataStore().ContactRequestReferenceReset(ctx); err != nil {
		return nil, errcode.ErrOrbitDBAppend.Wrap(err)
	}

	_, shareableContact := accountGroup.MetadataStore().GetIncomingContactRequestsStatus()
	rdvSeed := []byte(nil)

	if shareableContact != nil {
		rdvSeed = shareableContact.PublicRendezvousSeed
	}

	return &protocoltypes.ContactRequestResetReference_Reply{
		PublicRendezvousSeed: rdvSeed,
	}, nil
}

// ContactRequestSend enqueues a new contact request to be sent
func (s *service) ContactRequestSend(ctx context.Context, req *protocoltypes.ContactRequestSend_Request) (_ *protocoltypes.ContactRequestSend_Reply, err error) {
	ctx, _, endSection := tyber.Section(ctx, s.logger, "Sending contact request")
	defer func() { endSection(err, "") }()

	s.logger.Debug("Contact request info", tyber.FormatStepLogFields(ctx, []tyber.Detail{}, tyber.WithJSONDetail("Request", req))...)

	shareableContact := req.Contact
	if shareableContact == nil {
		return nil, errcode.ErrInvalidInput
	}

	accountGroup := s.getAccountGroup()
	if accountGroup == nil {
		return nil, errcode.ErrGroupMissing
	}

	if _, err := accountGroup.MetadataStore().ContactRequestOutgoingEnqueue(ctx, shareableContact, req.OwnMetadata); err != nil {
		return nil, errcode.ErrOrbitDBAppend.Wrap(err)
	}

	return &protocoltypes.ContactRequestSend_Reply{}, nil
}

// ContactRequestAccept accepts a contact request
func (s *service) ContactRequestAccept(ctx context.Context, req *protocoltypes.ContactRequestAccept_Request) (_ *protocoltypes.ContactRequestAccept_Reply, err error) {
	ctx, _, endSection := tyber.Section(ctx, s.logger, "Accepting contact request")
	defer func() { endSection(err, "") }()

	pk, err := crypto.UnmarshalEd25519PublicKey(req.ContactPK)
	if err != nil {
		return nil, errcode.ErrDeserialization.Wrap(err)
	}

	accountGroup := s.getAccountGroup()
	if accountGroup == nil {
		return nil, errcode.ErrGroupMissing
	}

	if _, err := accountGroup.MetadataStore().ContactRequestIncomingAccept(ctx, pk); err != nil {
		return nil, errcode.ErrOrbitDBAppend.Wrap(err)
	}

	if err = s.groupDatastore.PutForContactPK(ctx, pk, s.deviceKeystore); err != nil {
		return nil, err
	}

	return &protocoltypes.ContactRequestAccept_Reply{}, nil
}

// ContactRequestDiscard ignores a contact request without informing the request sender
func (s *service) ContactRequestDiscard(ctx context.Context, req *protocoltypes.ContactRequestDiscard_Request) (_ *protocoltypes.ContactRequestDiscard_Reply, err error) {
	ctx, _, endSection := tyber.Section(ctx, s.logger, "Discarding contact request")
	defer func() { endSection(err, "") }()

	pk, err := crypto.UnmarshalEd25519PublicKey(req.ContactPK)
	if err != nil {
		return nil, errcode.ErrDeserialization.Wrap(err)
	}

	accountGroup := s.getAccountGroup()
	if accountGroup == nil {
		return nil, errcode.ErrGroupMissing
	}

	if _, err := accountGroup.MetadataStore().ContactRequestIncomingDiscard(ctx, pk); err != nil {
		return nil, errcode.ErrOrbitDBAppend.Wrap(err)
	}

	return &protocoltypes.ContactRequestDiscard_Reply{}, nil
}
