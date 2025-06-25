package service

import (
	"errors"
	"statistic_service/internal/model"
	"statistic_service/internal/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const maxWalletsPerUser = 10

type WalletService interface {
	CreateWallet(userID uuid.UUID, name string) (*model.Wallet, error)
	ListWallets(userID uuid.UUID) ([]model.Wallet, error)
	GetWalletByID(walletID, userID uuid.UUID) (*model.Wallet, error)
	UpdateWalletName(walletID, userID uuid.UUID, newName string) (*model.Wallet, error)
	DeleteWallet(walletID, userID uuid.UUID) error
	InviteUserToWallet(walletID, inviterID uuid.UUID, invitedUserEmail string) error
	GetWalletMembers(walletID, requesterID uuid.UUID) ([]model.WalletMember, error)
	UpdateMemberRole(walletID, adminID, memberID uuid.UUID, newRole string) error
	RemoveMemberFromWallet(walletID, removerID, memberToRemoveID uuid.UUID) error
	CreateTransactionInWallet(walletID, userID uuid.UUID, tx *model.Transaction) (*model.Transaction, error)
	GetTransactionsForWallet(walletID, userID uuid.UUID) ([]model.Transaction, error)
}

type walletService struct {
	walletRepo   repository.WalletRepository
	userRepo     repository.UserRepository        // Для поиска пользователей по email
	txRepo       repository.TransactionRepository // <-- ДОБАВЬТЕ ЭТО
	categoryRepo repository.CategoryRepository
	logger       *logrus.Logger
}

func NewWalletService(walletRepo repository.WalletRepository, userRepo repository.UserRepository, txRepo repository.TransactionRepository, categoryRepo repository.CategoryRepository, logger *logrus.Logger,
) WalletService {
	return &walletService{
		walletRepo:   walletRepo,
		userRepo:     userRepo,
		txRepo:       txRepo,       // <-- ДОБАВЬТЕ ЭТО
		categoryRepo: categoryRepo, // <-- ДОБАВЬТЕ ЭТО
		logger:       logger,
	}
}

// CreateWallet создает кошелек и проверяет лимит
func (s *walletService) CreateWallet(userID uuid.UUID, name string) (*model.Wallet, error) {
	count, err := s.walletRepo.CountByOwnerID(userID)
	if err != nil {
		s.logger.Errorf("Failed to count wallets for user %s: %v", userID, err)
		return nil, errors.New("failed to check wallet limit")
	}
	if count >= maxWalletsPerUser {
		return nil, errors.New("wallet limit reached")
	}

	wallet := &model.Wallet{
		ID:      uuid.New().String(), // Преобразуем UUID в строку
		OwnerID: userID.String(),     // И здесь тоже
		Name:    name,
	}

	if err := s.walletRepo.Create(wallet); err != nil {
		s.logger.Errorf("Failed to create wallet for user %s: %v", userID, err)
		return nil, errors.New("could not create wallet")
	}

	s.logger.Infof("Wallet '%s' created for user %s", wallet.ID, userID)
	return wallet, nil
}

// ListWallets возвращает все кошельки, в которых пользователь является участником.
func (s *walletService) ListWallets(userID uuid.UUID) ([]model.Wallet, error) {
	wallets, err := s.walletRepo.ListByUserID(userID)
	if err != nil {
		s.logger.Errorf("Failed to list wallets for user %s: %v", userID, err)
		return nil, errors.New("could not retrieve wallets")
	}

	s.logger.Infof("Successfully listed %d wallets for user %s", len(wallets), userID)
	return wallets, nil
}

// GetWalletByID находит кошелек по ID и проверяет, что пользователь является его участником.
func (s *walletService) GetWalletByID(walletID, userID uuid.UUID) (*model.Wallet, error) {
	// Проверяем, является ли пользователь участником кошелька
	_, err := s.walletRepo.GetMember(walletID, userID)
	if err != nil {
		s.logger.Warnf("User %s attempted to access wallet %s without being a member", userID, walletID)
		return nil, errors.New("wallet not found or access denied")
	}

	wallet, err := s.walletRepo.GetByID(walletID)
	if err != nil {
		s.logger.Errorf("Could not retrieve wallet %s: %v", walletID, err)
		return nil, errors.New("could not retrieve wallet")
	}

	return wallet, nil
}

// UpdateWalletName обновляет имя кошелька. Только 'admin' может это делать.
func (s *walletService) UpdateWalletName(walletID, userID uuid.UUID, newName string) (*model.Wallet, error) {
	member, err := s.walletRepo.GetMember(walletID, userID)
	if err != nil {
		s.logger.Warnf("User %s attempted to update wallet %s without being a member", userID, walletID)
		return nil, errors.New("wallet not found or access denied")
	}

	// Только админ может менять имя кошелька
	if member.Role != model.WalletRoleAdmin {
		s.logger.Warnf("User %s (role: %s) attempted to update wallet %s without admin rights", userID, member.Role, walletID)
		return nil, errors.New("permission denied: only admins can update wallet name")
	}

	walletToUpdate, err := s.walletRepo.GetByID(walletID)
	if err != nil {
		return nil, errors.New("could not retrieve wallet for update")
	}

	walletToUpdate.Name = newName
	if err := s.walletRepo.Update(walletToUpdate); err != nil {
		s.logger.Errorf("Failed to update wallet %s: %v", walletID, err)
		return nil, errors.New("could not update wallet")
	}

	return walletToUpdate, nil
}

// DeleteWallet удаляет кошелек. Только владелец (owner) может это делать.
func (s *walletService) DeleteWallet(walletID, userID uuid.UUID) error {
	wallet, err := s.walletRepo.GetByID(walletID)
	if err != nil {
		s.logger.Warnf("User %s attempted to delete non-existent wallet %s", userID, walletID)
		return errors.New("wallet not found")
	}

	// Сравниваем ID пользователя из токена с ID владельца кошелька
	// Важно: в модели OwnerID - это string, а userID у нас uuid.UUID. Приводим к одному типу.
	if wallet.OwnerID != userID.String() {
		s.logger.Warnf("User %s attempted to delete wallet %s owned by %s", userID, walletID, wallet.OwnerID)
		return errors.New("permission denied: only the owner can delete the wallet")
	}

	// В репозитории метод Delete также проверяет owner_id для безопасности
	if err := s.walletRepo.Delete(walletID, userID); err != nil {
		s.logger.Errorf("Failed to delete wallet %s by owner %s: %v", walletID, userID, err)
		return errors.New("could not delete wallet")
	}

	s.logger.Infof("Wallet %s was deleted by owner %s", walletID, userID)
	return nil
}

// InviteUserToWallet приглашает нового пользователя в кошелек.
// Только администратор кошелька может приглашать.
func (s *walletService) InviteUserToWallet(walletID, inviterID uuid.UUID, invitedUserEmail string) error {
	// 1. Проверяем, что приглашающий - админ кошелька
	inviterMember, err := s.walletRepo.GetMember(walletID, inviterID)
	if err != nil {
		return errors.New("inviter is not a member of this wallet")
	}
	if inviterMember.Role != model.WalletRoleAdmin {
		return errors.New("permission denied: only admins can invite users")
	}

	// 2. Находим приглашаемого пользователя по email
	invitedUser, err := s.userRepo.GetByEmail(invitedUserEmail)
	if err != nil {
		return errors.New("user with the specified email not found")
	}

	invitedUserID, _ := uuid.Parse(invitedUser.ID)

	// 3. Проверяем, не является ли пользователь уже участником
	_, err = s.walletRepo.GetMember(walletID, invitedUserID)
	if err == nil {
		return errors.New("user is already a member of this wallet")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Если произошла другая ошибка, кроме "не найдено"
		return errors.New("failed to check membership")
	}

	// 4. Добавляем нового участника с ролью 'member'
	newMember := &model.WalletMember{
		WalletID: walletID.String(),
		UserID:   invitedUser.ID,
		Role:     model.WalletRoleMember,
	}

	if err := s.walletRepo.AddMember(newMember); err != nil {
		s.logger.Errorf("Failed to add member %s to wallet %s: %v", invitedUser.ID, walletID, err)
		return errors.New("could not add user to the wallet")
	}

	s.logger.Infof("User %s invited user %s to wallet %s", inviterID, invitedUser.ID, walletID)
	return nil
}

// GetWalletMembers возвращает список участников кошелька.
// Любой участник кошелька может просмотреть список.
func (s *walletService) GetWalletMembers(walletID, requesterID uuid.UUID) ([]model.WalletMember, error) {
	// Проверяем, что запрашивающий является участником кошелька
	_, err := s.walletRepo.GetMember(walletID, requesterID)
	if err != nil {
		return nil, errors.New("wallet not found or access denied")
	}

	members, err := s.walletRepo.GetMembers(walletID)
	if err != nil {
		s.logger.Errorf("Failed to get members for wallet %s: %v", walletID, err)
		return nil, errors.New("could not retrieve wallet members")
	}

	return members, nil
}

// UpdateMemberRole изменяет роль участника. Только админ может менять роли.
func (s *walletService) UpdateMemberRole(walletID, adminID, memberID uuid.UUID, newRole string) error {
	// 1. Проверяем, что тот, кто меняет роль - админ
	adminMember, err := s.walletRepo.GetMember(walletID, adminID)
	if err != nil {
		return errors.New("action performer is not a member of this wallet")
	}
	if adminMember.Role != model.WalletRoleAdmin {
		return errors.New("permission denied: only admins can change roles")
	}

	// 2. Нельзя менять свою роль самому себе этим методом
	if adminID == memberID {
		return errors.New("cannot change your own role")
	}

	// 3. Проверяем, что новый статус валиден
	if newRole != model.WalletRoleAdmin && newRole != model.WalletRoleMember {
		return errors.New("invalid role specified")
	}

	// 4. Нельзя понизить в правах владельца кошелька
	wallet, err := s.walletRepo.GetByID(walletID)
	if err != nil {
		return errors.New("could not find wallet")
	}
	if wallet.OwnerID == memberID.String() {
		return errors.New("cannot change the owner's role")
	}

	// 5. Обновляем роль
	if err := s.walletRepo.UpdateMemberRole(walletID, memberID, newRole); err != nil {
		s.logger.Errorf("Failed to update role for user %s in wallet %s: %v", memberID, walletID, err)
		return errors.New("failed to update role")
	}

	s.logger.Infof("User %s updated role for user %s in wallet %s to '%s'", adminID, memberID, walletID, newRole)
	return nil
}

// RemoveMemberFromWallet удаляет участника из кошелька. Только админ может удалять.
func (s *walletService) RemoveMemberFromWallet(walletID, removerID, memberToRemoveID uuid.UUID) error {
	// 1. Проверяем, что удаляющий - админ
	removerMember, err := s.walletRepo.GetMember(walletID, removerID)
	if err != nil {
		return errors.New("action performer is not a member of this wallet")
	}
	if removerMember.Role != model.WalletRoleAdmin {
		return errors.New("permission denied: only admins can remove members")
	}

	// 2. Нельзя удалить самого себя
	if removerID == memberToRemoveID {
		return errors.New("cannot remove yourself from the wallet")
	}

	// 3. Нельзя удалить владельца кошелька
	wallet, err := s.walletRepo.GetByID(walletID)
	if err != nil {
		return errors.New("could not find wallet")
	}
	if wallet.OwnerID == memberToRemoveID.String() {
		return errors.New("cannot remove the wallet owner")
	}

	// 4. Удаляем участника
	if err := s.walletRepo.RemoveMember(walletID, memberToRemoveID); err != nil {
		s.logger.Errorf("Failed to remove user %s from wallet %s: %v", memberToRemoveID, walletID, err)
		return errors.New("failed to remove member")
	}

	s.logger.Infof("User %s removed user %s from wallet %s", removerID, memberToRemoveID, walletID)
	return nil
}

// CreateTransactionInWallet создает транзакцию в кошельке.
func (s *walletService) CreateTransactionInWallet(walletID, userID uuid.UUID, tx *model.Transaction) (*model.Transaction, error) {
	// 1. Проверяем, что пользователь - участник кошелька
	_, err := s.walletRepo.GetMember(walletID, userID)
	if err != nil {
		return nil, errors.New("wallet not found or access denied")
	}

	// 2. Проверяем, что категория принадлежит пользователю, который создает транзакцию
	categoryID, _ := uuid.Parse(tx.CategoryID)
	_, err = s.categoryRepo.GetCategoryByID(categoryID, userID)
	if err != nil {
		return nil, errors.New("category not found or does not belong to the user")
	}

	// 3. Устанавливаем правильные ID
	tx.UserID = userID.String()
	tx.WalletID = walletID.String()

	// 4. Создаем транзакцию через основной репозиторий транзакций
	if err := s.txRepo.Create(tx); err != nil {
		s.logger.Errorf("Failed to create transaction in wallet %s by user %s: %v", walletID, userID, err)
		return nil, errors.New("failed to create transaction")
	}

	s.logger.Infof("User %s created transaction %s in wallet %s", userID, tx.ID, walletID)
	return tx, nil
}

// GetTransactionsForWallet возвращает транзакции для кошелька.
func (s *walletService) GetTransactionsForWallet(walletID, userID uuid.UUID) ([]model.Transaction, error) {
	// 1. Проверяем, что пользователь - участник кошелька
	_, err := s.walletRepo.GetMember(walletID, userID)
	if err != nil {
		return nil, errors.New("wallet not found or access denied")
	}

	// 2. Получаем транзакции из репозитория кошельков
	transactions, err := s.walletRepo.GetTransactionsByWalletID(walletID, nil, nil, "")
	if err != nil {
		s.logger.Errorf("Failed to get transactions for wallet %s: %v", walletID, err)
		return nil, errors.New("could not retrieve transactions")
	}

	return transactions, nil
}
